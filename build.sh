#!/bin/bash
rm -rf build/node*
make clean
make build-linux
touch build/CustomLogFlag
docker-compose down
docker network prune -y
make localnet-start
