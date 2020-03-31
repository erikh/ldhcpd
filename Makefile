IMAGE_NAME ?= ldhcpd:testing
CODE_PATH ?= /go/src/code.hollensbe.org/erikh/ldhcpd
GO_TEST := sudo go test -v ./... -race -count 1

DOCKER_CMD := docker run -it \
	--cap-add NET_ADMIN \
	-e IN_DOCKER=1 \
	-e SETUID=$$(id -u) \
	-e SETGID=$$(id -g) \
	-w $(CODE_PATH) \
	-v ${PWD}/.go-cache:/tmp/go-build-cache \
	-v ${PWD}:$(CODE_PATH) \
	$(IMAGE_NAME)

install:
	go install -v ./...

shell: build
	mkdir -p .go-cache
	$(DOCKER_CMD)	

build: get-box
	box -t $(IMAGE_NAME) box.rb

docker-check:
	@if [ -z "$${IN_DOCKER}" ]; then echo "You really don't want to do this"; exit 1; fi

clean-interfaces: docker-check
	(for i in veth0 veth1 veth2 veth3; do sudo ip link del $$i; done) || :
	sudo ip link set br0 down || :
	sudo brctl delbr br0 || :

interfaces: docker-check clean-interfaces
	sudo brctl addbr br0
	sudo ip link add type veth
	sudo brctl addif br0 veth0
	sudo ip link add type veth
	sudo brctl addif br0 veth2
	sudo ip link set br0 up
	for i in veth0 veth1 veth2 veth3; do sudo ip link set dev $$i up; done
	sudo ip addr add dev veth1 10.0.20.1/24

get-box:
	@if [ ! -f "$(shell which box)" ]; \
	then \
		echo "Need to install box to build the docker images we use. Requires root access."; \
		curl -sSL box-builder.sh | sudo bash; \
	fi

test:
	if [ -z "$${IN_DOCKER}" ]; then make build && $(DOCKER_CMD) $(GO_TEST); else $(GO_TEST); fi

.PHONY: test
