#!/usr/bin/env bash
set -e

IMAGE_NAME=wukong-go
IMAGE_URL=registry.cn-hongkong.aliyuncs.com/rallets/$IMAGE_NAME
docker build . -t $IMAGE_NAME
docker tag $IMAGE_NAME $IMAGE_URL
docker push $IMAGE_URL
