#!/bin/bash

FE_HOST="172.20.80.2"
FE_QUERY_PORT="9030"
BE_HOST="172.20.80.3"
BE_HEARTBEAT_PORT="9050"

echo "Checking Doris FE ($FE_HOST:$FE_QUERY_PORT) status..."

# 1. 优雅地等待 FE 端口开放
# while ! nc -z $FE_HOST $FE_QUERY_PORT; do
#   echo "Waiting for Doris FE to start..."
#   sleep 3
# done
while ! (echo > /dev/tcp/$FE_HOST/$FE_QUERY_PORT) >/dev/null 2>&1; do
  echo "Waiting for Doris FE to start via /dev/tcp..."
  sleep 3
done

echo "FE is up! Checking if BE is already registered..."

# 2. 检查 BE 是否已存在，避免重复报错
# 使用 -N 掉表头，-s 开启静默模式
BE_EXISTS=$(mysql -h$FE_HOST -P$FE_QUERY_PORT -uroot -N -s -e "SHOW BACKENDS" | grep "$BE_HOST")

if [ -z "$BE_EXISTS" ]; then
    echo "Registering BE ($BE_HOST:$BE_HEARTBEAT_PORT)..."
    mysql -h$FE_HOST -P$FE_QUERY_PORT -uroot -e "ALTER SYSTEM ADD BACKEND '$BE_HOST:$BE_HEARTBEAT_PORT';"
    echo "BE registered successfully."
else
    echo "BE is already in the cluster, skipping registration."
fi

echo "Init job finished."