FROM ubuntu:22.04

ENV LC_ALL=C.UTF-8 \
    TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive

# 设置时区
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list \
    && sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list \
    && apt update \
    && apt install -y --no-install-recommends tzdata \
    && ln -fs /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && apt clean \
    && rm -rf /var/lib/apt/lists/*

# Init
RUN apt update \
    && apt install -y \
    curl vim net-tools iputils-ping nmap masscan --fix-missing

# Install chrome instead of chromium-browser(can't do screenshot,why?)
RUN curl -LO https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && apt install -y ./google-chrome-stable_current_amd64.deb --fix-missing \
    && rm google-chrome-stable_current_amd64.deb
# Install nemo server
COPY . /opt/nemo
COPY docker/server.yml docker/worker.yml /opt/nemo/conf/
# Install nemo worker
RUN mkdir /opt/nemo_worker && mkdir /opt/nemo_worker/logs && mkdir /opt/nemo_worker/third_party
COPY daemon_worker_linux_amd64 /opt/nemo_worker/
