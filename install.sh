#!/usr/bin/env bash

wget https://olivetree.oss-cn-hangzhou.aliyuncs.com/logdog.tar.gz
tar xzf logdog.tar.gz
if [[ ! -d /etc/logdog/ ]]; then
    mkdir /etc/logdog/
fi
mv logdog /usr/local/bin/
mv logdog.toml /etc/logdog/
