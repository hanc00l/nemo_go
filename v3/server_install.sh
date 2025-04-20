#!/bin/bash

sudo apt-get update && sudo apt-get upgrade -y
#echo "Install MongoDB 8.0"
sudo apt-get install -y gnupg curl vim
curl -fsSL https://www.mongodb.org/static/pgp/server-8.0.asc | sudo gpg -o /usr/share/keyrings/mongodb-server-8.0.gpg --dearmor
echo "deb [arch=amd64,arm64 signed-by=/usr/share/keyrings/mongodb-server-8.0.gpg] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/8.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-8.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org
sudo systemctl enable --now mongod
sudo systemctl status mongod | grep "active (running)"

# 安装 Redis
sudo apt-get install -y redis
sudo systemctl enable redis-server.service
sudo systemctl start redis-server.service
sudo systemctl status redis-server.service | grep "active (running)"

#导入mongodb初始数据
mongosh docker/mongo-init.js

echo "install completed"