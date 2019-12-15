#!/bin/bash -xe

export SEMVER=$(cat ./.version)
export CONFIG_PKG="github.com/orbs-network/trash-panda/config"
export GIT_COMMIT=$(git rev-parse HEAD)

time GOOS=linux GOARCH=amd64 go build -ldflags "-w -extldflags '-static' -X $CONFIG_PKG.SemanticVersion=$SEMVER -X $CONFIG_PKG.CommitVersion=$GIT_COMMIT" -o _bin/trash-panda -a ./main.go
