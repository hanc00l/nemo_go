#!/usr/bin/env bash

rm -rf release/*
rm -f server_darwin_amd64 worker_darwin_amd64 server_linux_amd64 worker_linux_amd64 daemon_worker_darwin_amd64 daemon_worker_linux_amd64
echo > log/runtime.log
echo > log/access.log


