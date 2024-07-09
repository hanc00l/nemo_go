#!/bin/bash

echo "start to apt-update && apt upgrade..."
sudo apt-get update && sudo apt upgrade -y
sudo apt-get install vim mysql-server rabbitmq-server --fix-missing -y

# 开始初始化数据库，创建数据库、用户和表
read -r -p "start to init mysql database [Y/n]: " sure
if [ -z "$sure" ];then
    sure="y"
fi
case "$sure" in
    y|Y)
       if [ -f "docker/mysql/initdb.d/nemo.sql" ];then
         echo "check docker/mysql/initdb.d/nemo.sql exist..."
         else
         echo "check docker/mysql/initdb.d/nemo.sql not exist,exit..."
         exit
       fi
       read -r -p "input database name(default is nemo): " db_name
       if [ -z "$db_name" ];then
           db_name="nemo"
       fi
       read -r -p "input database user(default is nemo): " db_user
       if [ -z "$db_user" ];then
           db_user="nemo"
       fi
       read -r -p "input database password(default is nemo2020): " db_password
       if [ -z "$db_password" ];then
         db_password="nemo2020"
       fi
       sudo mysql -e "CREATE DATABASE $db_name DEFAULT CHARACTER SET utf8mb4;"
       sudo mysql -e "CREATE USER '$db_user'@'%' IDENTIFIED BY '$db_password';"
       sudo mysql -e "GRANT ALL PRIVILEGES ON $db_name.* TO '$db_user'@'%';FLUSH PRIVILEGES;"
       sudo mysql "$db_name" < docker/mysql/initdb.d/nemo.sql
       ;;
    n|N)
        echo "skip..."
        ;;
esac
# 开始初始化rabbitmq，创建用户名和授权
read -r -p "start to init rabbitmq [Y/n]: " sure
if [ -z "$sure" ];then
    sure="y"
fi
case "$sure" in
    y|Y)
      read -r -p "input rabbitmq user(default is nemo): " rabbitmq_user
           if [ -z "$rabbitmq_user" ];then
               rabbitmq_user="nemo"
           fi
           read -r -p "input rabbitmq password(default is nemo2020): " rabbitmq_password
           if [ -z "$rabbitmq_password" ];then
             rabbitmq_password="nemo2020"
           fi
      sudo rabbitmqctl add_user "$rabbitmq_user" "$rabbitmq_password"
      sudo rabbitmqctl set_permissions -p "/" "$rabbitmq_user" ".*" ".*" ".*"
       ;;
    n|N)
        echo "skip..."
        ;;
esac

# 生成随机的authKey
authKey=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 20)
echo "new authKey is:$authKey"
# 替换配置文件中的authKey
echo "set authKey for server and worker..."
sed -i "s/authKey: ZduibTKhcbb6Pi8W/authKey: $authKey/g" conf/server.yml
sed -i "s/authKey: ZduibTKhcbb6Pi8W/authKey: $authKey/g" conf/worker.yml
# 替换mysql数据训、用户名和密码：
echo "set database for server..."
# 替换指定19-21行的内容
sed -i "19s/name: nemo/name: $db_name/" conf/server.yml
sed -i "20s/username: nemo/username: $db_user/" conf/server.yml
sed -i "21s/password: nemo2020/password: $db_password/" conf/server.yml
# 替换rabbitmq的用户名和密码：
echo "set rabbitmq user and pass for server and worker..."
sed -i "s/username: guest/username: $rabbitmq_user/" conf/server.yml
sed -i "s/password: guest/password: $rabbitmq_password/" conf/server.yml
sed -i "s/username: guest/username: $rabbitmq_user/" conf/worker.yml
sed -i "s/password: guest/password: $rabbitmq_password/" conf/worker.yml

echo "server install finished!"
# server安装完成后，必须手动修改conf/worker.yml，设置rpc、filesync和rabbitmq的host ip，在worker同步后会覆盖worker的conf/worker.yml
# 如果server端没有修改worker.yml文件会导致同步后worker无法正常工作
echo "YOU MUST SET host ip (instead of 127.0.0.1 or localhost) for rpc、filesync and rabbitmq in conf/worker.yml !"



