#!/usr/bin/env bash
set -e

bin/build.sh
docker run -i -t -p 80:8080 wukong-go
