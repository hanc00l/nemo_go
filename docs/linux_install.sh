#!/bin/bash

#Tested on ubuntu18.04 LTS

#apt-get
sudo apt-get update \
    && sudo apt-get install wget curl vim net-tools git unzip python3-pip python3-setuptools \
    mysql-server rabbitmq-server \
    nmap whatweb masscan chromium-browser --fix-missing

# pip package
python3 -m pip install -U pip -i https://mirrors.aliyun.com/pypi/simple/ --user \
    && python3 -m pip install -U requests pocsuite3 -i https://mirrors.aliyun.com/pypi/simple/

# init databse and rabbitmq
sudo service mysql start \
    && mysql -u root -e 'CREATE DATABASE `nemo` DEFAULT CHARACTER SET utf8mb4;' \
    && mysql -u root -e 'CREATE USER "nemo"@"%" IDENTIFIED BY "nemo2020";GRANT ALL PRIVILEGES ON nemo.* TO "nemo"@"%";FLUSH PRIVILEGES;' \
    && mysql -u root nemo < docs/nemo.sql \

