#!/bin/sh

# start nemo server
cd /opt/nemo || exit
nohup ./server_linux_amd64 --web --service --cron &
# wait for nemo server to start
sleep 3s
# start nemo daemon worker
cd /opt/nemo_worker || exit
nohup ./daemon_worker_linux_amd64 --service 127.0.0.1 --auth da8Ae0e295eba72a7rgb1d34b13d86b --no-redis-proxy &
# keep running...
cd /opt || exit
tail -f /dev/null

