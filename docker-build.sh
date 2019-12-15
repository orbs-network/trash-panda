#!/bin/bash -xe

./build-binaries.sh

export SEMVER=$(cat ./.version)
docker build -t orbsnetwork/trash-panda:$SEMVER .
