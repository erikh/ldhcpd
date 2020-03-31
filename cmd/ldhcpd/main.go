package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"code.hollensbe.org/erikh/ldhcpd/dhcpd"
	"github.com/krolaw/dhcp4"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Action = serve

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func installSignalHandler(appName string, handler *dhcpd.Handler) {
	sigChan := make(chan os.Signal, 1)
	go func() {
		for {
			switch <-sigChan {
			// FIXME add config reload as SIGUSR1 or SIGHUP
			case syscall.SIGTERM, syscall.SIGINT:
				fmt.Printf("Stopping %v...", appName)
				handler.Close()
				fmt.Println("done.")
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

	handler, err := dhcpd.NewHandler(ctx.Args()[0], ctx.Args()[1])
	if err != nil {
		return errors.Wrap(err, "while configuring dhcpd")
	}

	installSignalHandler(ctx.App.Name, handler)

	return dhcp4.ListenAndServeIf(ctx.Args()[0], handler)
}
