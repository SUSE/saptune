#!/bin/sh

#echo "zypper in ..."
#zypper -n --gpg-auto-import-keys ref && zypper -n --gpg-auto-import-keys in go1.10 go rpcbind cpupower uuidd polkit tuned sysstat

/bin/systemctl start tuned
tuned-adm profile balanced

systemctl --no-pager status
# try to resolve systemd status 'degraded'
systemctl reset-failed
systemctl --no-pager status

echo "PATH is $PATH, GOPATH is $GOPATH, CI_TST_HOME is $CI_TST_HOME"

export CI_TST_HOME=/home/ci_tst
mkdir -p ${CI_TST_HOME}/gopath/src/github.com/SUSE
cd ${CI_TST_HOME}/gopath/src/github.com/SUSE
if [ ! -f saptune ]; then
	ln -s /app saptune
fi
export GO111MODULE=off
export GOPATH=${CI_TST_HOME}/gopath
export PATH=${CI_TST_HOME}/gopath/bin:$PATH
export CI_TST_BUILD_DIR=${CI_TST_HOME}/gopath/src/github.com/SUSE/saptune

echo "PATH is $PATH, GOPATH is $GOPATH, CI_TST_HOME is $CI_TST_HOME"
echo "ls -l /etc/saptune/*"
ls -l /etc/saptune/*

mkdir -p /etc/saptune/override
mkdir -p /var/lib/saptune/working
#if [ ! -d /usr/share/saptune/sols ]; then
if [ ! -d /var/lib/saptune/working/sols ]; then
	ln -s /app/testdata/sol/sols /var/lib/saptune/working/sols
fi

echo "go environment:"
go env
go version

cd saptune
pwd
ls -al

# to get TasksMax settings work, needs a user login session
echo "start nobody login session in background"
su --login nobody -c "sleep 4m" &
sleep 10
ps -ef
loginctl --no-pager

echo "exchange /etc/os-release"
cp /etc/os-release /etc/os-release_OrG

# for some sysctl tests
echo "vm.pagecache_limit_ignore_dirty = 1" > /etc/sysctl.d/saptune_test.conf
echo "vm.pagecache_limit_ignore_dirty = 1" > /etc/sysctl.d/saptune_test2.conf

echo "run go tests"
go test -v -coverprofile=c.out -cover ./...
exitErr=$?
go build
ps -ef
pkill -P $$
exit $exitErr
