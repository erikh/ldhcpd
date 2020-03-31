package main

import (
	"fmt"
	"os"

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

func serve(ctx *cli.Context) error {
	if len(ctx.Args()) != 2 {
		return errors.New("usage: dhcpd [interface] [config file]")
	}

	handler, err := dhcpd.NewHandler(ctx.Args()[0], ctx.Args()[1])
	if err != nil {
		return errors.Wrap(err, "while configuring dhcpd")
	}

	return dhcp4.ListenAndServeIf(ctx.Args()[0], handler)
}
