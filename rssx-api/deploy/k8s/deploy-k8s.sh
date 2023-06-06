#!/bin/sh
projectName="rssx-api"

cd /home/wiloon/projects/rssx/rssx-api || exit
CGO_ENABLED=1 GOOS=linux GOPROXY=https://goproxy.io go build -v -a -o ${projectName} ${projectName}.go
ls -l /home/wiloon/projects/rssx/rssx-api/rssx-api
sudo buildah bud -t registry.wiloon.com/rssx-api:v0.0.1 .
sudo buildah push registry.wiloon.com/rssx-api:v0.0.1
echo "y"|sudo podman image prune
rm /home/wiloon/projects/rssx/rssx-api/rssx-api
