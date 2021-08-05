#!/bin/sh

# start
cd /opt/nemo
nohup ./server_linux_amd64 &
nohup ./worker_linux_amd64 &
tail -f log/*.log

# keep running...
/bin/bash
