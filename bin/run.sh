#!/usr/bin/env bash
set -e

bin/build.sh
docker run -i -t -p 8080:80 wukong-go
