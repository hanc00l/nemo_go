#!/bin/bash

#Tested on ubuntu18.04 LTS

#apt-get
#只有server需要安装 mysql-server和rabbitmq-server
#只有worker需要安装 nmap masscan chromium-browser python及pocsuite
sudo apt-get update \
    && sudo apt-get install wget curl vim net-tools git unzip \
    mysql-server rabbitmq-server \
    python3-pip python3-setuptools nmap masscan chromium-browser --fix-missing

# pip package
# 只有worker需要安装python
python3 -m pip install -U pip -i https://mirrors.aliyun.com/pypi/simple/ --user \
    && python3 -m pip install -U requests pocsuite3 -i https://mirrors.aliyun.com/pypi/simple/

# init databse and rabbitmq
# 只有server需要
sudo service mysql start \
    && mysql -u root -e 'CREATE DATABASE `nemo` DEFAULT CHARACTER SET utf8mb4;' \
    && mysql -u root -e 'CREATE USER "nemo"@"%" IDENTIFIED BY "nemo2020";GRANT ALL PRIVILEGES ON nemo.* TO "nemo"@"%";FLUSH PRIVILEGES;' \
    && mysql -u root nemo < ../docker/mysql/initdb.d/docs/nemo.sql \

