#!/usr/bin/env bash
set -e

cd "$(dirname "$BASH_SOURCE")"

# Downloads dependencies into vendor/ directory
mkdir -p vendor
cd vendor

git_clone() {
	pkg=$1
	bra=$2

	pkg_url=https://$pkg
	target_dir=src/$pkg

	echo "$pkg @ $bra: "

	if [ -d $target_dir ]; then
		echo "rm old, $pkg"
		rm -fr $target_dir
	fi

	echo "clone, $pkg"
	git clone --depth 1 --quiet --branch $bra  $pkg_url $target_dir
	echo "done"
}

go_get() {
	pkg=$1

	echo "go get $pkg"
	GOPATH=`pwd` go get $pkg
	echo "done"
}


git_clone github.com/docker/libcontainer master

git_clone github.com/docker/docker v1.4.1
rm -rf src/github.com/docker/docker/vendor/src/github.com/docker/libcontainer #avoiding double dependency


echo "don't forget to add vendor folder to your GOPATH (export GOPATH=\$GOPATH:\`pwd\`/vendor)"
