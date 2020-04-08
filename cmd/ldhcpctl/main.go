package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"code.hollensbe.org/erikh/ldhcpd/proto"
	"code.hollensbe.org/erikh/ldhcpd/version"
	"github.com/erikh/go-transport"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	// Author is me
	Author = "Erik Hollensbe <erik+git@hollensbe.org>"
)

func main() {
	app := cli.NewApp()
	app.Name = "ldhcpctl"
	app.Usage = "Control ldhcpd"
	app.Version = version.Version
	app.Author = Author

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, t",
			Usage: "Set the host:port connection for GRPC",
			Value: "localhost:7846",
		},
		cli.StringFlag{
			Name:  "cert, c",
			Usage: "Set the client certificate for authentication",
			Value: "/etc/ldhcpd/client.pem",
		},
		cli.StringFlag{
			Name:  "key, k",
			Usage: "Set the client certificate key",
			Value: "/etc/ldhcpd/client.key",
		},
		cli.StringFlag{
			Name:  "ca",
			Usage: "Set the certificate authority",
			Value: "/etc/ldhcpd/rootCA.pem",
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
			Usage:     "Set a lease. `leasetime` is 'persistent', or a golang duration: https://golang.org/pkg/time/#ParseDuration",
			Description: `
Examples:

	ldhcpctl set 00:00:00:00:00:00 1.2.3.4 8m # expires in 8 minutes
	ldhcpctl set 00:00:00:00:00:00 1.2.3.4 persistent # renews until deleted
			`,
			Action: set,
			Flags: []cli.Flag{
				cli.DurationFlag{
					Name:  "grace-period, gp",
					Usage: "Grace period between lease expiration and hard reclaim time",
					Value: 8 * time.Hour,
				},
			},
		},
		{
			Name:      "list",
			ArgsUsage: "",
			Usage:     "List all leases in the table",
			Action:    list,
		},
		{
			Name:      "remove",
			ArgsUsage: "[mac address]",
			Usage:     "Remove a lease by mac address",
			Action:    remove,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getClient(ctx *cli.Context) (proto.LeaseControlClient, error) {
	cert, err := transport.LoadCert(ctx.GlobalString("ca"), ctx.GlobalString("cert"), ctx.GlobalString("key"), "")
	if err != nil {
		return nil, errors.Wrap(err, "while loading client certificate")
	}

	cc, err := transport.GRPCDial(cert, ctx.GlobalString("host"))
	if err != nil {
		return nil, errors.Wrap(err, "while configuring grpc client")
	}

	return proto.NewLeaseControlClient(cc), nil
}

func listLeases(leases []*proto.Lease) {
	/// func NewWriter(output io.Writer, minwidth, tabwidth, padding int, padchar byte, flags uint) *Writer {
	w := tabwriter.NewWriter(os.Stdout, 8, 2, 2, ' ', 0)
	w.Write([]byte(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n", "MAC", "IP", "Dynamic", "Persistent", "Lease End", "Grace Period End")))
	for _, lease := range leases {
		w.Write([]byte(fmt.Sprintf("%s\t%s\t%v\t%v\t%s\t%s\n", lease.MACAddress, lease.IPAddress, lease.Dynamic, lease.Persistent, time.Unix(lease.LeaseEnd.Seconds, 0), time.Unix(lease.LeaseGraceEnd.Seconds, 0))))
	}
	w.Flush()
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

	listLeases([]*proto.Lease{lease})
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

	var (
		leaseEnd   time.Duration
		persistent bool
	)

	leaseDuration := strings.TrimSpace(ctx.Args()[2])
	switch {
	case strings.HasPrefix(leaseDuration, "persist"):
		leaseEnd = time.Hour
		persistent = true
	default:
		var err error
		leaseEnd, err = time.ParseDuration(ctx.Args()[2])
		if err != nil {
			return errors.Wrap(err, "while parsing lease end duration")
		}
	}

	_, err = client.SetLease(context.Background(), &proto.Lease{
		MACAddress:    ctx.Args()[0],
		IPAddress:     ctx.Args()[1],
		Persistent:    persistent,
		LeaseEnd:      &timestamp.Timestamp{Seconds: time.Now().Add(leaseEnd).Unix()},
		LeaseGraceEnd: &timestamp.Timestamp{Seconds: time.Now().Add(ctx.Duration("grace-period")).Unix()},
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

	listLeases(leases.List)

	return nil
}

func remove(ctx *cli.Context) error {
	if len(ctx.Args()) != 1 {
		return errors.New("invalid arguments")
	}

	client, err := getClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.RemoveLease(context.Background(), &proto.MACAddress{Address: ctx.Args()[0]})
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %s\n", ctx.Args()[0])
	return nil
}
