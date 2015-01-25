#sc-redis

Super fast and easy way to deploy (containerized) redis-server:

````bash
$  sc-redis
[Stage 0] pid 1007
[Stage 0] exporting redis container rootfs
[Stage 1] bridge scredis0 up 10.0.5.0/8
[Stage 1] container IP address: 10.0.5.2/8
[Stage 2] pid 1 (inside container)
[Stage 2] container is in  /tmp/demo2/scredis_20150125200856
[Stage 2] executing [redis-server /etc/redis.conf]

[...] #redis-server output

$ redis-cli -h 10.0.5.2 -p 6379
10.0.5.2:6379> ping
PONG
````

##Description

`sc-redis` is **dependency free**, with just the binary you will be able to spawn self contained redis-server instances.

`sc-redis` uses [libcontainer](https://github.com/docker/libcontainer), the go library used as container backend in docker.
A minimal redis-server image is built and packaged directly inside `sc-redis` binary with [go-bindata](https://github.com/jteeuwen/go-bindata).
On start, `sc-redis` extract the image (rootfs), create a container with libcontainer and run
redis-server in it.

You can read more about [**how the image is built**](https://github.com/robinmonjo/sc-redis/blob/master/BUILD_IMAGE.md)

Every `sc-redis` containers will be hooked on the network bridge `scredis0` created on
the host. Each `sc-redis` process is containerized, totally isolated from the host
system or from other running `sc-redis` process.

For now, redis-server container are ephemeral and will be destroyed when the process exits.
Do not use it if you need storage persistence.

##Installation

````bash
curl -sL https://github.com/robinmonjo/sc-redis/releases/download/v0.1/sc-redis-v0.1_x86_64.tgz | tar -C /usr/local/bin -zxf -
````

To uninstall:
* remove the binary `/usr/local/bin/sc-redis`
* remove the ip tracker file `/etc/scredis_ips.json`
* delete the bridge iface: `ip link delete scredis0 type bridge`

##Usage

`sudo sc-redis` launch the server and output its ip address. You can then connect to it directly:
`redis-cli -h 10.0.5.XXX -p 6379`

If you want the server to be accessible outside of the host, you will need some iptables magic.

####flags

- `-c "config line, config line"`

Allows to pass custom [redis-server configuration](http://redis.io/topics/config). Each configuration, separated by a ","
will be written at the end of the [default `redis.conf`](https://raw.githubusercontent.com/antirez/redis/2.8/redis.conf) file.

Example: `sc-redis -c "requirepass foobar, port 9999"`

- `-v`

Display `sc-redis` version. Sample output:

`sc-redis v0.1 (redis v2.8.19, libcontainer v1.4.0)`

##Roadmap

- [ ] allow user to choose container ip address
- [ ] allow to directly use the host net interface (and not `scredis0` bridge)
- [ ] start using an existing rootfs (data persistence)

##Hacking on sc-redis

The Makefile contains a lot of info but basically, to get started:

1. fork this repository and clone it
2. `make vendor`
3. `make redis-rootfs` (as it's not versionned)
4. `make build` done !

Note, if you are working on Vagrant, running `sc-redis` on a shared folder won't work (rootfs extraction will fail).

##Credits

* libcontainer [guys](https://github.com/docker/libcontainer/blob/master/MAINTAINERS) for their amazing work
* inspired by [@crosbymichael's tweet](https://twitter.com/crosbymichael/status/543235554263830528)

##License

MIT
