# some hints for development

the sources should be available at $GOPATH/src/github.com/SUSE/saptune

## build saptune v2
	cd $GOPATH/src/github.com/SUSE/saptune
	go build

## lint and format checks for the sources before committing changes

	gofmt -d *
	golint ./...
	go vet -composites=false ./...

and run the unit tests (in a docker container)

## unit tests for saptune:
after committing the changes to git travis is used for automatic testing

But before committing the sources, run the tests locally by using docker and the same workflow as on travis

	su -
	systemctl start docker
	cd $GOPATH/src/github.com/SUSE/saptune
	docker run --name travis-st-ci --privileged -v /sys/fs/cgroup:/sys/fs/cgroup:ro -td -v "$(pwd):/app" shap/crmsh
	docker exec -t travis-st-ci /bin/sh -c "cd /app; ./run_travis_tst.sh;"

in $GOPATH/src/github.com/SUSE/saptune

	go tool cover -html=c.out -o coverage.html

and check the file 'coverage.html' in your Browser to see the coverage

make changes to the source files

and run the tests again

	docker exec -t travis-st-ci /bin/sh -c "cd /app; ./run_travis_tst.sh;"

clean up when finished with your tests

	docker stop travis-st-ci
	docker rm travis-st-ci

