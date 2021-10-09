FROM ubuntu:18.04

ENV LC_ALL C.UTF-8

# Init
RUN set -x \
    # You may need this if you're in Mainland China
    && sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list \
    ###
    && apt-get update \
    && apt-get install -y  python3-pip python3-setuptools \
    wget curl vim net-tools  iputils-ping git unzip \
    nmap masscan chromium-browser --fix-missing

# pip package
RUN set -x \
    # You may need this if you're in Mainland China
    # && python3.7 -m pip config set global.index-url 'https://pypi.mirrors.ustc.edu.cn/simple/' \
    && python3 -m pip install -U pip -i https://mirrors.aliyun.com/pypi/simple/ --user \
    && python3 -m pip install -U requests pocsuite3 -i https://mirrors.aliyun.com/pypi/simple/

COPY . /opt/nemo

#link to mysql and rabbitmq
RUN set -x \
    && sed -i 's/host: 127.0.0.1/host: mysql/g' /opt/nemo/conf/server.yml \
    && sed -i 's/host: localhost/host: rabbitmq/g' /opt/nemo/conf/server.yml \
    && sed -i 's/host: localhost/host: rabbitmq/g' /opt/nemo/conf/worker.yml
