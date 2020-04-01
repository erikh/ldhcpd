## light dhcpd

This is a dhcpd with very few features.

## Instructions

- make test: This runs the tests.
- make shell: This runs a docker shell. You can do a few things in here:
  - make test: This is context-dependent and will run properly in the container
  - make interfaces: This sets up some dummy interfaces and plumbs them through
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

## Author

Erik Hollensbe <erik+git@hollensbe.org>
