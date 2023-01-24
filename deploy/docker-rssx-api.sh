#!/bin/sh
projectName="rssx-api"
version="v0.0.1"
cd ~/projects/rssx/rssx-api || exit

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.io go build -v -a -o ${projectName} ${projectName}.go
ls -l ~/projects/rssx/rssx-api/rssx-api
sudo buildah bud --arch=amd64 -t registry.wiloon.com/rssx-api:${version}-amd64 .
sudo buildah push registry.wiloon.com/rssx-api:${version}-amd64
rm ~/projects/rssx/rssx-api/rssx-api

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 GOPROXY=https://goproxy.io go build -v -a -o ${projectName} ${projectName}.go
ls -l ~/projects/rssx/rssx-api/rssx-api
sudo buildah bud --arch=arm64 -t registry.wiloon.com/rssx-api:${version}-arm64 .
sudo buildah push registry.wiloon.com/rssx-api:${version}-arm64
rm ~/projects/rssx/rssx-api/rssx-api

buildah manifest create registry.wiloon.com/rssx-api:${version} \
    --amend registry.wiloon.com/rssx-api:${version}-amd64 \
    --amend registry.wiloon.com/rssx-api:${version}-arm64

buildah manifest inspect registry.wiloon.com/rssx-api:${version}
buildah manifest push --all registry.wiloon.com/rssx-api:${version}  docker://registry.wiloon.com/rssx-api:${version}

echo "y"|sudo podman image prune
rm ~/projects/rssx/rssx-api/rssx-api

#ansible -i '192.168.50.100,' all -u root -m shell -a 'podman pull registry.wiloon.com/rssx-api:v0.0.1'
#ansible -i '192.168.50.100,' all -u root -m shell -a 'podman stop rssx-api'
#ansible -i '192.168.50.100,' all -u root -m shell -a 'podman rm rssx-api'
#ansible -i '192.168.50.100,' all -u root -m shell -a 'podman run -d --name rssx-api -p 3000:8080 -v /etc/localtime:/etc/localtime:ro -v rssx-api-config:/etc/rssx-api -v rssx-api-log:/var/log/rssx-api -v rssx-api-data:/var/lib/rssx-api registry.wiloon.com/rssx-api:v0.0.1'
##ansible -i '192.168.50.100,' all -u root -m shell -a 'podman image prune'
#ansible -i '192.168.50.100,' all -u root -m shell -a 'podman ps -f name=rssx-api'
