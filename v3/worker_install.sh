#!/bin/bash

sudo apt-get update && sudo apt-get upgrade -y
sudo apt-get install curl vim net-tools iputils-ping nmap masscan --fix-missing -y
# set nmap and masscan to be run as root
nemo_pwd=$(pwd)
cd /usr/bin && sudo chown root nmap masscan && sudo chmod u+s nmap masscan
cd "$nemo_pwd" || exit
# install google chrome
curl -L https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb -o /tmp/google-chrome-stable_current_amd64.deb \
 && sudo apt install -y /tmp/google-chrome-stable_current_amd64.deb \
 && rm /tmp/google-chrome-stable_current_amd64.deb

echo "install completed"