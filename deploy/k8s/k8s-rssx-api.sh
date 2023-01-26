#!/bin/sh

projectName="rssx-api"
version="v0.0.1"

cd ~/projects/rssx/rssx-api || exit

# go-sqlite3 requires cgo
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.io go build -v -a -o ${projectName} ${projectName}.go
ls -l ~/projects/rssx/rssx-api/rssx-api
md5sum ~/projects/rssx/rssx-api/rssx-api
sudo buildah bud --arch=amd64 -t registry.wiloon.com/rssx-api:${version}-amd64 .
sudo buildah push registry.wiloon.com/rssx-api:${version}-amd64
rm ~/projects/rssx/rssx-api/rssx-api

CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc CC_FOR_TARGET=gcc-aarch64-linux-gnu GOPROXY=https://goproxy.io go build -v -a -o ${projectName} ${projectName}.go
ls -l ~/projects/rssx/rssx-api/rssx-api
md5sum ~/projects/rssx/rssx-api/rssx-api
sudo buildah bud --arch=arm64 -t registry.wiloon.com/rssx-api:${version}-arm64 .
sudo buildah push registry.wiloon.com/rssx-api:${version}-arm64
rm ~/projects/rssx/rssx-api/rssx-api

podman image ls
podman manifest rm registry.wiloon.com/rssx-api:${version}
podman image ls

buildah manifest create registry.wiloon.com/rssx-api:${version} \
    --amend registry.wiloon.com/rssx-api:${version}-amd64 \
    --amend registry.wiloon.com/rssx-api:${version}-arm64

buildah manifest inspect registry.wiloon.com/rssx-api:${version}
buildah manifest push --all registry.wiloon.com/rssx-api:${version}  docker://registry.wiloon.com/rssx-api:${version}

echo "y"|sudo podman image prune

ansible -i '192.168.50.228,' all -u root -m copy -a 'src=~/projects/rssx/deploy/k8s/k8s-rssx-api-deployment.yaml dest=/tmp'
ansible -i '192.168.50.228,' all -u root -m shell -a 'kubectl delete -f /tmp/k8s-rssx-api-deployment.yaml'
ansible -i '192.168.50.228,' all -u root -m shell -a 'kubectl create -f /tmp/k8s-rssx-api-deployment.yaml'