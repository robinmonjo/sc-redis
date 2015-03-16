#How the redis-server image is built

Prerequisite: [`krgo`](https://github.com/robinmonjo/krgo) must be installed. All the operations are made on an
ubuntu 14.04 box

* Build redis on a ubuntu 14.04 host

````bash
$ wget http://download.redis.io/releases/redis-2.8.19.tar.gz
$ tar xzf redis-2.8.19.tar.gz
$ cd redis-2.8.19
$ make
````

* pull the busybox linux image and move the redis-server binary into it:

````bash
$ sudo krgo pull busybox:ubuntu-14.04 -r busybox -g
$ cd busybox
$ sudo mkdir -p usr/local/bin
$ sudo cp /home/vagrant/redis-2.8.19/src/redis-server ./usr/local/bin/
````

The diff between the original busybox image and the one we are building for `sc-redis` is simply:

````bash
$ git status
On branch layer_3_f6169d24347d30de48e4493836bec15c78a34f08cc7f17d6a45a19d68dc283ac
Changes to be committed:
  (use "git reset HEAD <file>..." to unstage)

	new file:   usr/local/bin/redis-server
````

Then this image is sent on the docker hub using `krgo`.
The resulting image weights around 9 MB and is available [here](https://registry.hub.docker.com/u/robinmonjo/scredis/)
