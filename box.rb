from "golang:1.14"

run "apt update && apt install bridge-utils isc-dhcp-client sudo -y"
env GOCACHE: "/tmp/go-build-cache"
run %q[grep -vE 'env_reset|secure_path' /etc/sudoers >tmp && mv tmp /etc/sudoers]
run %q[echo 'username ALL=(ALL:ALL) NOPASSWD:ALL' >>/etc/sudoers]
copy "entrypoint.sh", "/entrypoint.sh"
run "chmod 755 /entrypoint.sh"
set_exec entrypoint: ["sh", "/entrypoint.sh"], cmd: ["bash"]
