package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"code.hollensbe.org/erikh/ldhcpd/dhcpd"
	"code.hollensbe.org/erikh/ldhcpd/proto"
	"code.hollensbe.org/erikh/ldhcpd/version"
	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func main() {
	app := cli.NewApp()

	app.Version = version.Version

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "disable, d",
			Usage: "disable the DHCPd DHCP offering (control-plane only)",
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

	config, err := dhcpd.ParseConfig(ctx.Args()[1])
	if err != nil {
		return errors.Wrap(err, "while parsing configuration")
	}

	db, err := config.NewDB()
	if err != nil {
		return errors.Wrap(err, "while initialiing database")
	}

	handler, err := dhcpd.NewHandler(ctx.Args()[0], config, db)
	if err != nil {
		return errors.Wrap(err, "while configuring dhcpd")
	}

	srv := proto.Boot(db)
	l, err := net.Listen("tcp", "localhost:7846")
	if err != nil {
		return errors.Wrap(err, "while configuring grpc listener")
	}
	installSignalHandler(ctx.App.Name, srv, l, handler)

	go srv.Serve(l)

	if ctx.GlobalBool("disable") {
		select {} // will never reach dhcp listen
	}

	return dhcp4.ListenAndServeIf(ctx.Args()[0], handler)
}
