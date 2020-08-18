#!/bin/bash

set -e
set -x

TESTS_TYPE=$1
ARGS="${@:2}"

function stop_and_cleanup() {
  docker-compose logs checker
  docker-compose logs writer
  docker-compose down
}

trap stop_and_cleanup SIGINT SIGTERM EXIT

if [ "${TESTS_TYPE}" == "e2e" ]; then
  docker-compose run --rm wait-infrastructure
fi

docker-compose run --rm ${ARGS} test_${TESTS_TYPE}
