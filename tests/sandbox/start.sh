#!/bin/bash

ROOT_DIR=$(pwd -P)
PATH="$PATH:$(dirname $0)"

$(which gsed || which sed) -i "/ssh-key-path/assh-key-owner: $USER" ${ROOT_DIR}/tests/sandbox/config.yml
export TEST_CONFIG=${ROOT_DIR}/tests/sandbox/config.yml

pushd ${ROOT_DIR}
    go test \
        -coverpkg="github.com/shatteredsilicon/ssm-manage/..." \
        -c -tags testrunmain -o ./ssm-configurator.test \
        ./cmd/ssm-configurator
popd

exec ${ROOT_DIR}/ssm-configurator.test \
    -test.run "^TestRunMain$" \
    -test.coverprofile=coverage.txt
