##!/bin/ash

set -e

sudo apt-get update -qq

echo "Installing base stack"

packagelist=(
	cgroup-lite             #this is important !!
	git
  curl
  golang
)

sudo DEBIAN_FRONTEND=noninteractive apt-get install -y ${packagelist[@]}

curl -sL https://github.com/robinmonjo/cargo/releases/download/v1.4.1/cargo-v1.4.1_x86_64.tgz | tar -C /usr/local/bin -zxf -

echo "GOPATH=~/code/go" >> ~/.bashrc
