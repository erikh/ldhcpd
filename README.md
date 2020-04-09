## Light DHCPd

This is a DHCP service/daemon with very few features. It provides basic dynamic
IPv4 pool allocation as well as persistent, static leases. iPXE support does
not exist yet, but is planned.

One thing Light DHCPd offers that is novel, is a remote control plane powered
over GRPC, authenticated and encrypted by TLS (ecdsa certs are only supported
currently) client certificates. This control plane can be embedded into your
orchestration code or you can use the provided command-line tool to manipulate
it from your shell. There is a golang client, and the protobufs are included
in the source tree if you wish to generate clients for other languages.

There is a small configuration file for managing dynamic leases and the overall
network parameters (resolver, gateway).

## Stability

Light DHCPd has been powering my network for over a week!

**(Seriously, you do not want to use this in production yet.)**

## Development Instructions

- `make shell`: This runs a docker shell. You can do a few things in here:
  - `make test`: This is context-dependent and will run properly in the container
    or outside of it, running the unit and functional tests in a container.
  - `make interfaces`: This sets up some dummy interfaces and plumbs them through
    a bridge; afterwards `veth1` will be available for running a `ldhcpd` on, and
    `veth3` will be available for running a `dhclient` on.
  - `make start`: will start the dhcpd.
  - `make stop`: stops it.
  - `make get-ip`: issues a ISC `dhclient` launch against the second veth pair
    to get an IP, allowing you to test interaction with a real client.

**NOTE**: The following instructions install https://github.com/box-builder/box
on first run to build the images, will require `sudo` for that (but nothing
else). If you don't want to run `sudo` to install box, install it in your
`$PATH` somewhere and try again.

Assuming you're not crazy enough to try this on your own network, try this at
your shell instead:

```bash
$ make shell
# <inside of container>
$ make interfaces
$ make install
$ sudo ldhcpd veth1 example.conf &
$ sudo dhclient -1 -v -d veth3
# ^C it to stop it
$ ip addr
# veth3 -> 10.0.20.50
```

If you want to boot the control plane only, without serving DHCP, try the `-d`
flag.

## Config File Rundown

```yaml
#
# DNS servers
#
dns_servers:
  - 10.0.0.1
  - 1.1.1.1

#
# network gateway
#
gateway: 10.0.20.1

#
# Dynamic Range of IPs to use in dynamic lease hand-outs, IP inclusive.
#
dynamic_range:
  from: 10.0.20.50
  to: 10.0.20.100

#
# Lease parameters:
#
# The duration is the duration of the lease; no other allocation can affect the
# IP you will get back while this lease is obtained.
#
# The grace period is the maximum amount of time the IP is available to the mac
# address; it is added to the duration. If another mac comes in and there are
# no available IPs, addresses in # the grace period may be reclaimed to make
# room.
#
lease:
  duration: 24h
  grace_period: 8h
```

## Dependencies

- https://github.com/krolaw/dhcp4 for the dhcp4 protocol work, thanks to
  Richard Warburton (@krolaw), et al. This tool would be much less useful
  without it.
- https://github.com/jinzhu/gorm and https://github.com/mattn/go-sqlite3 for the database work.
- https://google.golang.org/grpc for the control plane protocol
- https://github.com/box-builder/box which is a mruby-based docker image builder.

## Author

Erik Hollensbe <erik+git@hollensbe.org>
