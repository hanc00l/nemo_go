FROM ubuntu:18.04

ENV LC_ALL C.UTF-8

# Init
RUN set -x \
    # You may need this if you're in Mainland China
    && sed -i 's/archive.ubuntu.com/mirrors.ustc.edu.cn/g' /etc/apt/sources.list \
    ###
    && apt-get update \
    && apt-get install -y  python3-pip python3-setuptools \
    wget curl vim net-tools git unzip \
    mysql-server rabbitmq-server \
    nmap whatweb masscan chromium-browser --fix-missing

# pip package
RUN set -x \
    # You may need this if you're in Mainland China
    # && python3.7 -m pip config set global.index-url 'https://pypi.mirrors.ustc.edu.cn/simple/' \
    && python3 -m pip install -U pip -i https://mirrors.aliyun.com/pypi/simple/ --user \
    && python3 -m pip install -U requests pocsuite3 -i https://mirrors.aliyun.com/pypi/simple/
    
COPY . /opt/nemo
# init databse and rabbitmq
RUN set -x \
    && service mysql start \
    && mysql -u root -e 'CREATE DATABASE `nemo` DEFAULT CHARACTER SET utf8mb4;' \
    && mysql -u root -e 'CREATE USER "nemo"@"%" IDENTIFIED BY "nemo2020";GRANT ALL PRIVILEGES ON nemo.* TO "nemo"@"%";FLUSH PRIVILEGES;' \
    && mysql -u root nemo < /opt/nemo/docs/nemo.sql \
    && cp /opt/nemo/docs/docker_start.sh /opt/nemo \
    && chmod +x /opt/nemo/docker_start.sh \

WORKDIR /opt/nemo
ENTRYPOINT ["/opt/nemo/docker_start.sh"]
EXPOSE 5000
