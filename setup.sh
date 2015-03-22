##!/bin/ash

set -e

sudo apt-get update -qq

echo "Installing base stack"

packages=(
	cgroup-lite             #this is important !!
	git
  curl
)

sudo DEBIAN_FRONTEND=noninteractive apt-get install -y ${packages[@]}

curl -sL https://github.com/robinmonjo/krgo/releases/download/v1.5.0/krgo-v1.5.0_x86_64.tgz | tar -C /usr/local/bin -zxf -

#install latest go version
curl -sL https://storage.googleapis.com/golang/go1.4.1.linux-amd64.tar.gz | tar -C /usr/local/ -zxf -
echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
source /etc/profile

#install go-bindata
mkdir -p /home/vagrant/go
GOPATH=/home/vagrant/go go get -u github.com/jteeuwen/go-bindata/...
echo "export PATH=$PATH:/home/vagrant/go/bin" >> /etc/profile
