#!/usr/bin/env bash

echo "package nemo for macos and linux..."
echo "MAKS SURE \"RunMode = Release\" in pkg/conf/config.go"
rm -rf release/*
echo > log/runtime.log

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_darwin_amd64 cmd/server/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_darwin_amd64 cmd/worker/main.go
tar -cvzf release/nemo_darwin_amd64.tar \
  --exclude=thirdparty/xray/xray_linux_amd64 \
  server_darwin_amd64 worker_darwin_amd64 \
  conf log thirdparty web docs

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_linux_amd64 cmd/server/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_linux_amd64 cmd/worker/main.go
tar -cvzf release/nemo_linux_amd64.tar \
  --exclude=thirdparty/xray/xray_darwin_amd64 \
  server_linux_amd64 worker_linux_amd64 \
  Dockerfile conf log thirdparty web docs

rm -f server_darwin_amd64 worker_darwin_amd64 server_linux_amd64 worker_linux_amd64

echo "package done..."
