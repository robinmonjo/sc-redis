# sc-redis

Super fast and easy way to deploy (containerized) redis-server:

````bash
$  sc-redis
[host] pid 4840
[host] container uid sc_redis_2d24808
[host] exporting redis container rootfs
[host] writing redis configuration
[container] pid 1
[container] starting redis

[...] #redis-server output

$ redis-cli
127.0.0.1:6379> ping
PONG
````

## Description

`sc-redis` (**s**elf **c**ontained or **s**tatic **c**ontainer) is **dependency free**, with just the binary you will be able to spawn self contained redis-server instances.

`sc-redis` uses [libcontainer](https://github.com/docker/libcontainer), the go library used as container backend in docker.
A minimal redis-server image is built and packaged directly inside `sc-redis` binary with [go-bindata](https://github.com/jteeuwen/go-bindata).
On start, `sc-redis` extract the image (rootfs), create a container with libcontainer and run
redis-server in it.

You can read more about [**how the image is built**](https://github.com/robinmonjo/sc-redis/blob/master/BUILD_IMAGE.md)

Each `sc-redis` process is containerized, totally isolated from the host system or from other running `sc-redis` process.

For now, redis-server container are ephemeral and will be destroyed when the process exits.
Do not use it if you need storage persistence.

## Installation

````bash
curl -sL https://github.com/robinmonjo/sc-redis/releases/download/v1.1.2/sc-redis-v1.1.2_x86_64.tgz | tar -C /usr/local/bin -zxf -
````

To uninstall:
* remove the binary `/usr/local/bin/sc-redis`
* delete the bridge iface: `ip link delete scredis0 type bridge` (if you used `-i` flag)

## Usage

`sudo sc-redis [-v] [-i 172.18.xxx.xxx] [-c "redis conf, redis conf, redis conf"] [-w working_directory]`


#### flags

- `-i 172.18.<xxx>.<xxx>`

If this flag is used, the container uses the net namespace and is accessible through the *scredis0* bridge automatically created on the host.
You can then connect to it this way: `redis-cli -h 172.18.<xxx>.<xxx> -p 6379`.

If you want the server to be accessible outside of the host, you will need some iptables magic.

Note: it is user's responsibility to make sure there is no ip conflict


- `-c "config line, config line"`

Allows to pass custom [redis-server configuration](http://redis.io/topics/config). Each configuration, separated by a ","
will be written at the end of the [default `redis.conf`](https://raw.githubusercontent.com/antirez/redis/2.8/redis.conf) file.

Example: `sc-redis -c "requirepass foobar, port 9999"`

- `-w working_directory`

Directory where to extract container rootfs. Current working directory by default.

- `-v`

Display `sc-redis` version. Sample output:

`sc-redis version sc-redis v1.0 (redis v2.8.19, libcontainer b6cf7a6c8520fd21e75f8b3becec6dc355d844b0)`

## Contributing

The Makefile contains a lot of info but basically, to get started:

1. fork this repository and clone it
2. `make vendor`
3. `make redis-rootfs` (as it's not versioned, you will need [krgo](https://github.com/robinmonjo/krgo) in your path)
4. `make build` done !

Note, if you are working on Vagrant, running `sc-redis` on a shared folder won't work (rootfs extraction will fail). You can run integration tests with `make test`.

Feel free to report any issues or improvement ideas if you have some.

I hope this project may show the way for other self contained apps. `sc-mongodb` anyone ? `sc-postgresql`, `sc-*` ?
One command deployment compensates for fat binaries ;)

## Credits

* libcontainer [guys](https://github.com/docker/libcontainer/blob/master/MAINTAINERS) for their amazing work on libcontainer and [nsinit](https://github.com/docker/libcontainer/tree/master/nsinit)
* inspired by [@crosbymichael's tweet](https://twitter.com/crosbymichael/status/543235554263830528)

## License

MIT
