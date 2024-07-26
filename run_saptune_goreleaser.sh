#!/bin/sh

echo "----------"
echo "Building SAPTUNE binary in $bdate with version $Version"
echo "----------"
zypper -n --gpg-auto-import-keys install git
git config --global --add safe.directory /app
go mod init github.com/SUSE/saptune
go mod tidy
go install github.com/goreleaser/goreleaser@v1.24.0
$(go env GOPATH)/bin/goreleaser release
