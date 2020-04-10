groupadd -g "$SETGID" groupname
useradd -d "$PWD" -u "$SETUID" -g "$SETGID" -G sudo username 
echo '===> Installing cert store *inside the container*'
sudo -EH -u username mkcert -install && \
  sudo -EH -u username mkcert -cert-file /etc/ldhcpd/server.pem -key-file /etc/ldhcpd/server.key localhost 127.0.0.1 && \
  sudo -EH -u username mkcert -client -cert-file /etc/ldhcpd/client.pem -key-file /etc/ldhcpd/client.key localhost 127.0.0.1
exec sudo -EH -u username "$@"
