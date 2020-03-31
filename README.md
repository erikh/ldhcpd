## light dhcpd

This is a dhcpd with very few features.

## Instructions

- make test: this runs the tests
- make shell: this runs a docker shell. you can do a few thing in here:
  - make test: this is context-dependent and will run properly in the container
  - make interfaces: this sets up some dummy interfaces and plumbs them through
    a bridge; afterwards veth1 will be available for running a ldhcpd on, and
    veth3 will be available for running a dhclient on.

Assuming you're not crazy enough to try this on your own network, try this at
your shell instead:

```bash
$ make shell
# <inside of container>
$ make interfaces
$ make install
$ sudo ldhcpd veth1 example.conf &
$ sudo dhclient -1 -v -d veth3
```

## Author

Erik Hollensbe <erik+git@hollensbe.org>
