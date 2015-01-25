#sc-redis

Super fast and easy way to deploy (containerized) redis-server:

![Alt text](https://dl.dropboxusercontent.com/u/6543817/sc-redis-readme/sc-redis.png)

##Usage

Running `sc-redis` will spawn a containerized redis-server attached to a network bridge

````bash
$> sudo sc-redis


````

###flags

- `-c "config line, config line"`

Allows to pass [redis-server configuration](http://redis.io/topics/config). Each configuration, separated by a ","
will be written at the end of the `redis.conf` file.

Example: `sc-redis -c "requirepass foobar, port 9999"`

- `-v`

Display `sc-redis` version. Sample output: `sc-redis v0.1 (redis v2.8.19, libcontainer v1.4.0)`

##How it works

`sc-redis` uses [libcontainer](https://github.com/docker/libcontainer), a go library used as container backend in docker.
A minimal redis-server image is build (see build image instruction) and packaged
directly inside `sc-redis` binary using [go-bindata](https://github.com/jteeuwen/go-bindata).
On start, `sc-redis` extract the image (rootfs), create a container with libcontainer and run
redis-server in it.

Every `sc-redis` containers will be hooked on the network bridge `scredis0` created on
the host. Each `sc-redis` process is containerized, meaning it's totally isolated from the host
system or from other running `sc-redis` process.

##Why ?


##Roadmap

- [ ] allow user to choose container ip address
- [ ] allow to directly use the host net interface (and not `scredis0` bridge)
- [ ] start using an existing rootfs (data persistence)

##Uninstall

* remove the binary `/usr/local/bin/sc-redis`
* remove the ip tracker file `/etc/scredis_ips.json`
* delete the sc-redis bridge iface: `ip link delete scredis0 type bridge`

##Credits

* libcontainer [guys](https://github.com/docker/libcontainer/blob/master/MAINTAINERS) for their amazing work
* inspired by [@crosbymichael's tweet](https://twitter.com/crosbymichael/status/543235554263830528)

##License

MIT
