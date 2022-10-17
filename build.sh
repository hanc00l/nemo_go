#!/usr/bin/env bash

echo "build nemo for macos,linux and windows..."

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_darwin_amd64 cmd/server/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_darwin_amd64 cmd/worker/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o daemon_worker_darwin_amd64 cmd/daemon_worker/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_linux_amd64 cmd/server/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_linux_amd64 cmd/worker/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o daemon_worker_linux_amd64 cmd/daemon_worker/main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o worker_windows_amd64.exe cmd/worker/main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o daemon_worker_windows_amd64.exe cmd/daemon_worker/main.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o server_windows_amd64.exe cmd/server/main.go


echo "build done."