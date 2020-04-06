from "golang:1.14"

PROTOC_VERSION = "3.11.4"
PROTOC_URL     = "https://github.com/protocolbuffers/protobuf/releases/download/v#{PROTOC_VERSION}/protoc-#{PROTOC_VERSION}-linux-x86_64.zip"

MKCERT_VERSION = "1.4.1"
MKCERT_URL = "https://github.com/FiloSottile/mkcert/releases/download/v#{MKCERT_VERSION}/mkcert-v#{MKCERT_VERSION}-linux-amd64"

def download(name, url)
  run "curl -sSL -o /#{name} '#{url}'"
  yield "/#{name}"
  run "rm -f /#{name}"
end

run "apt update && apt install bridge-utils isc-dhcp-client sudo sqlite3 unzip curl -y"
env GOCACHE: "/tmp/go-build-cache"
run %q[grep -vE 'env_reset|secure_path' /etc/sudoers >tmp && mv tmp /etc/sudoers]
run %q[echo 'username ALL=(ALL:ALL) NOPASSWD:ALL' >>/etc/sudoers]
copy "entrypoint.sh", "/entrypoint.sh"
run "chmod 755 /entrypoint.sh"

download("protoc.zip", PROTOC_URL) do |path|
  run "unzip #{path} -d /usr"
  run "chmod -R 755 /usr/bin/protoc /usr/include/google"
end

download("mkcert", MKCERT_URL) do |path|
  run "chmod 0755 '#{path}'"
  run "mv '#{path}' /usr/local/bin"
end

run "mkdir /etc/ldhcpd && chown 1000:1000 /etc/ldhcpd"
env CAROOT: "/etc/ldhcpd"

set_exec entrypoint: ["sh", "/entrypoint.sh"], cmd: ["bash"]
