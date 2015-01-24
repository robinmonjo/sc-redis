GOPATH:=`pwd`/vendor:$(GOPATH)
GOPATH:=`pwd`/vendor/src/github.com/docker/docker/vendor:$(GOPATH)
VERSION:=0.1
HARDWARE=$(shell uname -m)
DOCKER_IMAGE=robinmonjo/scredis

build: vendor
	GOPATH=$(GOPATH) go build

redis-rootfs:
	#need cargo in the path
	#need go-bindata in the path
	rm -f redis_rootfs.go
	cd /tmp && sudo cargo pull $(DOCKER_IMAGE) -r redis_rootfs #going to /tmp to make sure not in vagrant shared folder
	cd /tmp && sudo tar cf redis_rootfs.tar -C redis_rootfs .
	cd /tmp && go-bindata -o redis_rootfs.go -nomemcopy redis_rootfs.tar
	mv /tmp/redis_rootfs.go .

test: vendor
	GOPATH=$(GOPATH) go install
	GOPATH=$(GOPATH) go test

release:
	mkdir -p release

clean:
	rm -rf ./sc-redis ./release

vendor:
	sh vendor.sh
