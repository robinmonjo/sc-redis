GOPATH:=`pwd`/vendor:$(GOPATH)
GOPATH:=`pwd`/vendor/src/github.com/docker/libcontainer/vendor:$(GOPATH)
VERSION:=0.1
HARDWARE=$(shell uname -m)

build: vendor
	GOPATH=$(GOPATH) go build

test: vendor
	GOPATH=$(GOPATH) go install
	GOPATH=$(GOPATH) go test

release:
	mkdir -p release

clean:
	rm -rf ./sc-redis ./release

vendor:
	sh vendor.sh
