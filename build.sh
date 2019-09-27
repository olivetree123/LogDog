#!/usr/bin/env bash

if [[ -f logdog ]]; then
    echo "logdog exists, remove"
    rm logdog
fi

if [[ -f logdog.tar.gz ]]; then
    echo "logdog.tar.gz exists, remove"
    rm logdog.tar.gz
fi

go build -o logdog .
tar czf logdog.tar.gz logdog logdog.toml
