IMAGE_NAME ?= ldhcpd:testing
CODE_PATH ?= /go/src/code.hollensbe.org/erikh/ldhcpd

shell: build
	mkdir -p .go-cache
	docker run -it --cap-add NET_ADMIN -e SETUID=$$(id -u) -e SETGID=$$(id -g) -w $(CODE_PATH) -v ${PWD}/.go-cache:/tmp/go-build-cache -v ${PWD}:$(CODE_PATH) $(IMAGE_NAME)

build:
	box -t $(IMAGE_NAME) box.rb

clean-interfaces:
	(for i in veth0 veth1 veth2 veth3; do sudo ip link del $$i; done) || :
	sudo ip link set br0 down
	sudo brctl delbr br0 || :

interfaces: clean-interfaces
	sudo brctl addbr br0
	sudo ip link add type veth
	sudo brctl addif br0 veth0
	sudo ip link add type veth
	sudo brctl addif br0 veth2
	sudo ip link set br0 up
	for i in veth0 veth1 veth2 veth3; do sudo ip link set dev $$i up; done
	sudo ip addr add dev veth1 10.0.20.1/24
