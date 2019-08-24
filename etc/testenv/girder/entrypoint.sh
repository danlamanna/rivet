#!/bin/bash
set -e

until nc -z mongodb 27017; do sleep 1; done;
(girder serve --host 0.0.0.0 --database mongodb://mongodb:27017/girder > entrypoint.log 2>&1) &
until grep -qi 'engine bus started' entrypoint.log; do sleep 1; done;

python /bootstrap.py

kill "$(pgrep -f girder)"

girder serve --host 0.0.0.0 --database mongodb://mongodb:27017/girder
