#!/bin/bash

# Check if docker is installed
if ! [ -x "$(command -v docker)" ]; then
  echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
  exit 1
fi

# Build and deploy
echo "Building the faas-cloud-badge"
docker build -t s8sg/faas-cloud-badge:1.0.0 faas-cloud-badge/ --no-cache


echo "deploying functions"
faas-cli deploy -f stack.yml
