#!/bin/sh
set -e

if [ "$(id -u)" = "0" ]; then
    mkdir -p /app/data
    chown -R appuser:appuser /app/data
    exec su-exec appuser "$@"
fi

exec "$@"
