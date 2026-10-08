#!/bin/sh
set -eu
mkdir -p /data/runtime
export SYNCRIA_SUPERVISED=1
while :; do
  if [ -x /data/runtime/current ]; then
    binary=/data/runtime/current
  else
    binary=/usr/local/bin/syncria
  fi
  set +e
  "$binary" "$@"
  code=$?
  set -e
  if [ "$code" -ne 75 ]; then
    exit "$code"
  fi
done
