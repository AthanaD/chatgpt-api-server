#!/bin/bash
set -e
now=$(date +"%Y%m%d%H%M%S")
# 将 dev 版本推送到 latest
docker tag xyhelper/chatgpt-api-server:dev xyhelper/chatgpt-api-server:latest
docker push xyhelper/chatgpt-api-server:latest
# 删除 dev 版本 防止重复提交
docker rmi xyhelper/chatgpt-api-server:dev
# 以当前时间为版本
docker tag xyhelper/chatgpt-api-server:latest xyhelper/chatgpt-api-server:$now
docker push xyhelper/chatgpt-api-server:$now
echo "release success" $now
# 写入发布日志 release.log
echo $now >> ./release.log
