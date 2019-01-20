#!/usr/bin/env bash
#
# Starts/Stops a docker container with a mysql db. Helpful for developing/testing.
#

DATA_DIR=mysql_datadir
PW=my-secret-pw
DB_NAME=spaceschalter

usage() {
    echo "Usage $0 (start|stop|exec|create-schema|import-testdata)"
}

current_id() {
  local R_ID
  R_ID=$(docker ps --last 1 | grep 'mysql:5.7' | cut -c 1-12)
  if [ "${R_ID}" == "" ]; then
    echo "Please start the docker mysql."
    exit 1
  fi
  echo ${R_ID}
}

if [ "$1" == "" ]; then
    usage
    exit 1
fi

case "$1" in
  start)
    mkdir -p ${DATA_DIR}
    # https://hub.docker.com/_/mysql/
    docker run --rm -it -v $(pwd)/${DATA_DIR}:/var/lib/mysql \
      -p 3306:3306 \
      -e MYSQL_ROOT_PASSWORD=${PW} \
      mysql:5.7 \
      --default-authentication-plugin=mysql_native_password \
      --innodb_buffer_pool_instances=1 \
      --innodb_buffer_pool_size=5M \
      --innodb_log_buffer_size=256K \
      --innodb_log_file_size=2M \
      --key_buffer_size=8 \
      --innodb_file_per_table=0
    ;;
  stop)
    docker stop $(current_id)
    ;;
  exec)
    shift
    echo "$*" | docker exec -i $(current_id) mysql -u root --password="${PW}" ${DB_NAME}
    ;;
  create-schema)
    echo "CREATE DATABASE ${DB_NAME}" | docker exec -i $(current_id) mysql -u root --password="${PW}"
    cat db-schema.sql | docker exec -i $(current_id) mysql -u root --password="${PW}" ${DB_NAME}
    ;;
  import-testdata)
    cat test/db-testdata.sql | docker exec -i $(current_id) mysql -u root --password="${PW}" ${DB_NAME}
    ;;
  *)
    usage
    exit 1
esac