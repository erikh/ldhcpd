package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

const (
	// Version is the version of the program
	Version = "0.1.0"

	// Author is me
	Author = "Erik Hollensbe <erik+git@hollensbe.org>"
)

func main() {
	app := cli.NewApp()
	app.Name = "ldhcpctl"
	app.Usage = "Control ldhcpd"
	app.Version = Version
	app.Author = Author

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, t",
			Usage: "Set the host:port connection for GRPC",
			Value: "localhost:7846",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:      "get",
			ArgsUsage: "[mac address]",
			Usage:     "Get a lease based on the mac address provided",
			Action:    get,
		},
		{
			Name:      "set",
			ArgsUsage: "[mac address] [ip address] [leasetime]",
			Usage:     "Set a lease. `leasetime` is a golang duration: https://golang.org/pkg/time/#ParseDuration",
			Action:    set,
		},
		{
			Name:      "list",
			ArgsUsage: "",
			Usage:     "List all leases in the table",
			Action:    list,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getClient(ctx *cli.Context) (proto.LeaseControlClient, error) {
	// FIXME add security
	cc, err := grpc.Dial(ctx.GlobalString("host"), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "while configuring grpc client")
	}

	return proto.NewLeaseControlClient(cc), nil
}

func get(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("invalid arguments")
	}

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	lease, err := client.GetLease(context.Background(), &proto.MACAddress{Address: ctx.Args()[0]})
	if err != nil {
		return errors.Wrapf(err, "while obtaining lease for %v", ctx.Args()[0])
	}

	fmt.Println("mac", lease.MACAddress)
	fmt.Println("ip", lease.IPAddress)
	fmt.Println("dynamic", lease.Dynamic)
	fmt.Println("lease_end", lease.LeaseEnd.String())

	return nil
}

func set(ctx *cli.Context) error {
	if len(ctx.Args()) != 3 {
		return errors.New("invalid arguments")
	}

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	d, err := time.ParseDuration(ctx.Args()[2])
	if err != nil {
		return errors.Wrap(err, "while parsing lease end duration")
	}

	_, err = client.SetLease(context.Background(), &proto.Lease{
		MACAddress: ctx.Args()[0],
		IPAddress:  ctx.Args()[1],
		LeaseEnd:   &timestamp.Timestamp{Seconds: time.Now().Add(d).Unix()},
	})
	if err != nil {
		return errors.Wrap(err, "error during lease set")
	}

	return nil
}

func list(ctx *cli.Context) error {
	if len(ctx.Args()) != 0 {
		return errors.New("invalid arguments")
	}

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	leases, err := client.ListLeases(context.Background(), &empty.Empty{})
	if err != nil {
		return errors.Wrap(err, "could not list leases")
	}

	/// func NewWriter(output io.Writer, minwidth, tabwidth, padding int, padchar byte, flags uint) *Writer {
	w := tabwriter.NewWriter(os.Stdout, 8, 2, 2, ' ', 0)
	w.Write([]byte(fmt.Sprintf("%s\t%s\t%s\t%s\n", "MAC", "IP", "Dynamic", "Lease End")))
	for _, lease := range leases.List {
		w.Write([]byte(fmt.Sprintf("%s\t%s\t%v\t%s\n", lease.MACAddress, lease.IPAddress, lease.Dynamic, lease.LeaseEnd)))
	}
	w.Flush()

	return nil
}
