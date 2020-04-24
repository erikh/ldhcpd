package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/erikh/go-transport"
	"github.com/erikh/ldhcpd/dhcpd"
	"github.com/erikh/ldhcpd/proto"
	"github.com/erikh/ldhcpd/version"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func main() {
	app := cli.NewApp()

	app.Name = "ldhcpd"
	app.Usage = "Light DHCPd server"
	app.Author = version.Author
	app.Version = version.Version

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "disable, d",
			Usage: "disable the DHCPd DHCP offering (control-plane only)",
		},
		cli.StringFlag{
			Name:  "listen, l",
			Usage: "Change the host:port to listen for GRPC connections",
			Value: "localhost:7846",
		},
	}

	app.Action = serve

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func installSignalHandler(appName string, grpcS *grpc.Server, l net.Listener, handler *dhcpd.Handler) {
	sigChan := make(chan os.Signal, 1)
	go func() {
		for {
			switch <-sigChan {
			// FIXME add config reload as SIGUSR1 or SIGHUP
			case syscall.SIGTERM, syscall.SIGINT:
				logrus.Infof("Stopping %v...", appName)
				grpcS.GracefulStop()
				l.Close()
				handler.Close()
				logrus.Infof("Done.")
				os.Exit(0)
			}
		}
	}()
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
}

func serve(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return errors.Errorf("usage: %s [interface] [config file]", ctx.App.Name)
	}

	ip, err := dhcpd.InterfaceIP(ctx.Args()[0])
	if err != nil {
		return errors.Wrap(err, "while discovering interface IP")
	}

	config, err := dhcpd.ParseConfig(ctx.Args()[1])
	if err != nil {
		return errors.Wrap(err, "while parsing configuration")
	}

	db, err := config.NewDB()
	if err != nil {
		return errors.Wrap(err, "while initialiing database")
	}

	handler, err := dhcpd.NewHandler(ip, config, db)
	if err != nil {
		return errors.Wrap(err, "while configuring dhcpd")
	}

	cert, err := config.Certificate.NewCert()
	if err != nil {
		return errors.Wrap(err, "while configuring transport credentials")
	}

	srv := proto.Boot(db)
	l, err := transport.Listen(cert, "tcp", ctx.GlobalString("listen"))
	if err != nil {
		return errors.Wrap(err, "while configuring grpc listener")
	}
	installSignalHandler(ctx.App.Name, srv, l, handler)

	go srv.Serve(l)

	if ctx.GlobalBool("disable") {
		select {} // will never reach dhcp listen
	}

	laddr := &net.UDPAddr{Port: 67}

	server, err := server4.NewServer(ctx.Args()[0], laddr, handler.ServeDHCP, server4.WithDebugLogger())
	if err != nil {
		return errors.Wrap(err, "while creating DHCP service")
	}

	return server.Serve()
}
