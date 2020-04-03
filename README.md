## Light DHCPd

This is a DHCP service/daemon with very few features. It provides basic dynamic
pool allocation as well as persistent, static leases. iPXE support does not
exist but is planned.

One thing light DHCPd offers that is novel, is a remote control plane powered
over GRPC. This control plane can be embedded into your orchestration code or
you can use the provided command-line tool to manipulate it from your shell.
There is a golang client, and the protobufs are included in the source tree if
you wish to generate clients for other languages.

## Instructions

- `make shell`: This runs a docker shell. You can do a few things in here:
  - `make test`: This is context-dependent and will run properly in the container
    or outside of it, running the unit and functional tests in a container.
  - `make interfaces`: This sets up some dummy interfaces and plumbs them through
    a bridge; afterwards `veth1` will be available for running a `ldhcpd` on, and
    `veth3` will be available for running a `dhclient` on.

Assuming you're not crazy enough to try this on your own network, try this at
your shell instead:

```bash
# installs box to build the images, will require sudo for that (but nothing else)
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

## Dependencies

- https://github.com/box-builder/box which is a mruby-based docker image builder.
- https://github.com/krolaw/dhcp4 for the dhcp4 protocol work, thanks to the authors.
- https://github.com/jinzhu/gorm and https://github.com/mattn/go-sqlite3 for the database work.
- https://google.golang.org/grpc for the control plane protocol

## Author

Erik Hollensbe <erik+git@hollensbe.org>
