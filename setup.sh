##!/bin/ash

set -e

sudo add-apt-repository ppa:duh/golang
sudo apt-get update -qq

echo "Installing base stack"

packagelist=(
	cgroup-lite             #this is important !!
	git
  curl
)

sudo DEBIAN_FRONTEND=noninteractive apt-get install -y ${packagelist[@]}

curl -sL https://github.com/robinmonjo/cargo/releases/download/v1.4.1/cargo-v1.4.1_x86_64.tgz | tar -C /usr/local/bin -zxf -

#install latest go version
curl -sL https://storage.googleapis.com/golang/go1.4.1.linux-amd64.tar.gz | tar -C /usr/local -zxf -
echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile

echo "GOPATH=~/code/go" >> ~/.bashrc
