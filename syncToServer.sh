#!/bin/sh

SERVER=k_spacebus
SERVER_PATH=/home/status2/app

rsync -n -avzi --no-perms --no-owner --no-group --delete --exclude config.toml status2 webUI ${SERVER}:${SERVER_PATH}/
echo ""
echo "This was a DRY RUN"
echo "Press ENTER to sync or CTRL + C to cancel."
read k

rsync -avzi --no-perms --no-owner --no-group --delete --exclude config.toml status2 webUI ${SERVER}:${SERVER_PATH}/