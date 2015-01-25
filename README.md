#sc-redis

The fastest way to run (containerised) redis


###Uninstall
* remove the binary `/usr/local/bin/sc-redis`
* remove the ip tracker file `/etc/scredis_ips.json`
* delete the sc-redis bridge iface: `ip link delete scredis0 type bridge`

###Credits

* libcontainer [guys](https://github.com/docker/libcontainer/blob/master/MAINTAINERS)
* inspired by [@crosbymichael tweet](https://twitter.com/crosbymichael/status/543235554263830528)

###License

MIT
