#!/bin/sh

echo "zypper in ..."
#/bin/systemctl start dbus -> does not work any longer
# additional libs needed to get 'tuned' working
zypper -n --gpg-auto-import-keys ref && zypper -n --gpg-auto-import-keys in glib2 glib2-tools libgio-2_0-0 libglib-2_0-0 libgmodule-2_0-0 libgobject-2_0-0 go1.10 go rpcbind cpupower uuidd polkit tuned sysstat

# dbus can not be started directly, only by dependency - so start 'tuned' instead
/bin/systemctl start tuned
systemctl --no-pager status
# try to resolve systemd status 'degraded'
systemctl reset-failed
systemctl --no-pager status

echo "PATH is $PATH, GOPATH is $GOPATH, TRAVIS_HOME is $TRAVIS_HOME"

export TRAVIS_HOME=/home/travis
mkdir -p ${TRAVIS_HOME}/gopath/src/github.com/SUSE
cd ${TRAVIS_HOME}/gopath/src/github.com/SUSE
if [ ! -f saptune ]; then
	ln -s /app saptune
fi
export GOPATH=${TRAVIS_HOME}/gopath
export PATH=${TRAVIS_HOME}/gopath/bin:$PATH
export TRAVIS_BUILD_DIR=${TRAVIS_HOME}/gopath/src/github.com/SUSE/saptune

mkdir -p /etc/saptune/override
mkdir -p /usr/share/saptune
if [ ! -f /usr/share/saptune/solutions ]; then
	ln -s /app/testdata/saptune-test-solutions /usr/share/saptune/solutions
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

echo "run go tests"
go test -v -coverprofile=c.out -cover ./...
exitErr=$?
go build
ps -ef
pkill -P $$
exit $exitErr
