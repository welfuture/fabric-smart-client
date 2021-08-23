#!/bin/bash

#
# Copyright IBM Corp All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

DOCKER_IMAGE_NAME ?= dlt-interop/relay
DOCKER_TAG ?= latest
DOCKER_IMAGE_INTEROPCHAINCODE = $(DOCKER_IMAGE_NAME)-interopchaincode:$(DOCKER_TAG)

### Check if a directory does not exist ###
if [ ! -d ".build" ]
then
    echo "Directory .build DOES NOT exists, create it..."
    mkdir .build
fi

cd .build
rm -rf weaver
mkdir weaver
cd weaver
WEAVER_DIR=$(pwd)

echo Clone weaver-dlt-interoperability...
git clone https://github.com/hyperledger-labs/weaver-dlt-interoperability.git

echo Build server docker image...
cd $WEAVER_DIR/weaver-dlt-interoperability/core/relay
make build-server-local

echo Build Fabric driver docker image...
cd $WEAVER_DIR/weaver-dlt-interoperability/core/drivers/fabric-driver
cp .env.template .env
make build-image-local

echo Package Interop Chaincode...
docker build . -f images/Dockerfile.interopchaincode -t $(DOCKER_IMAGE_INTEROPCHAINCODE)

echo Cleanup...
cd $WEAVER_DIR/..
rm -rf weaver