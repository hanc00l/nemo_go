#!/bin/bash

rm worker_linux_amd64.tar

tar -cvzf worker_linux_amd64.tar \
  --exclude=thirdparty/xray/xray_darwin_amd64 --exclude=conf/app.conf --exclude=server.yml \
   worker_linux_amd64 conf log thirdparty version.txt

