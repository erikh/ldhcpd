groupadd -g "$SETGID" groupname
useradd -d "$PWD" -u "$SETUID" -g "$SETGID" -G sudo username 
exec sudo -EH -u username "$@"
