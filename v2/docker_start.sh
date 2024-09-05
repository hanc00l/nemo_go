#!/bin/sh

# start
cd /opt/nemo
if [ $# -eq 0 ]
    then
        nohup ./server_linux_amd64 &
        nohup ./daemon_worker_linux_amd64 &
        nohup ./thirdparty/goby/goby-cmd-linux -mode api -bind 127.0.0.1:8361 -apiauth goby:goby &
else
    if [ "$1" = "server" ]
        then
            nohup ./server_linux_amd64 &
    fi
    if [ "$1" = "worker" ]
        then
            nohup ./daemon_worker_linux_amd64 &
            nohup ./thirdparty/goby/goby-cmd-linux -mode api -bind 127.0.0.1:8361 -apiauth goby:goby &
    fi
fi
tail -f log/*.log

# keep running...
/bin/bash
