#!/bin/sh

project_name="rssx-api"
version="v0.0.1"

cd ~/projects/rssx/rssx-api || exit

# go-sqlite3 requires cgo
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOPROXY=https://athens.wiloon.com go build -v -a -o ${projectName} ${projectName}.go
ls -l ~/projects/rssx/rssx-api/rssx-api
md5sum ~/projects/rssx/rssx-api/rssx-api

# Set the required variables
export REGISTRY="registry.wiloon.com"

sudo podman image ls
sudo podman manifest rm registry.wiloon.com/rssx-api:${version}
sudo podman image ls

# Create a multi-architecture manifest
export manifest_name=${project_name}-manifest
sudo buildah manifest create ${manifest_name}:${version}

sudo buildah bud --arch=amd64 -t registry.wiloon.com/rssx-api:${version} --manifest ${manifest_name} .
rm ~/projects/rssx/rssx-api/rssx-api

CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CC_FOR_TARGET=gcc-aarch64-linux-gnu GOPROXY=https://athens.wiloon.com go build -v -a -o ${projectName} ${projectName}.go
ls -l ~/projects/rssx/rssx-api/rssx-api
md5sum ~/projects/rssx/rssx-api/rssx-api

sudo podman image ls

sudo buildah bud --arch=arm64 -t registry.wiloon.com/rssx-api:${version} --manifest ${manifest_name} .
rm ~/projects/rssx/rssx-api/rssx-api

sudo buildah manifest push --all \
    ${manifest_name} \
    docker://registry.wiloon.com/rssx-api:${version}

echo "y"|sudo podman image prune

ansible -i '192.168.50.228,' all -u root -m copy -a 'src=~/projects/rssx/deploy/k8s/k8s-rssx-api-deployment.yaml dest=/tmp'
ansible -i '192.168.50.228,' all -u root -m shell -a 'kubectl delete -f /tmp/k8s-rssx-api-deployment.yaml'
ansible -i '192.168.50.228,' all -u root -m shell -a 'kubectl create -f /tmp/k8s-rssx-api-deployment.yaml'