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

## build the saptune package:
saptune is build on ibs (and not obs, as saptune is not available on Factory)

branch from a maintained project (see https://maintenance.suse.de/maintained)
	osc -A https://api.suse.de bco -M SUSE:SLE-12-SP2:Update saptune

build the source archive from the github repository (something like 'tar -czvf ../saptune-\<release\>.tgz .') and move it to your obs directory
Or - if the new version is already created in github -
copy the archive from https://github.com/SUSE/saptune/releases to your obs directory

change the saptune.spec file, at least the version field

change the saptune.changes file.
	first line should be '- update version of saptune v2 to \<new version\>'
	Add a description of the changes and do not forget to add the bsc# or jsc# reference to these changes.
	And don't forget the line length restriction :-)
	Important - changes of SAP Notes need to be mentioned in the changes file

change the \_service file and add the new version number

change the \_servicedata and add the commit id

commit the changes to the obs sub project and check the build

test the resulting package (initial install and update installations) before submitting a maintenance request.

