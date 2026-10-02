#!/bin/bash
set -e

# load .env
export $(grep -v '^#' .env | xargs)

ACTION=${1:-up}

migrate -path migrations \
  -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" \
  $ACTION
