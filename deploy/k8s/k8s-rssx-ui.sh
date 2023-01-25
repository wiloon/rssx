#!/bin/bash

version="v0.0.1"

cd ~/projects/rssx/rssx-ui || exit
yarn install
export NODE_OPTIONS=--openssl-legacy-provider
yarn build

sudo buildah bud --arch=amd64 -t registry.wiloon.com/rssx-ui:${version}-amd64 .
sudo buildah push registry.wiloon.com/rssx-ui:${version}-amd64

sudo buildah bud --arch=arm64 -t registry.wiloon.com/rssx-ui:${version}-arm64 .
sudo buildah push registry.wiloon.com/rssx-ui:${version}-arm64

rm -rf ~/projects/rssx/rssx-ui/dist

buildah manifest create registry.wiloon.com/rssx-ui:${version} \
    --amend registry.wiloon.com/rssx-ui:${version}-amd64 \
    --amend registry.wiloon.com/rssx-ui:${version}-arm64

buildah manifest inspect registry.wiloon.com/rssx-ui:${version}
buildah manifest push --all registry.wiloon.com/rssx-ui:${version}  docker://registry.wiloon.com/rssx-ui:${version}


ansible -i '192.168.50.228,' all -u root -m copy -a 'src=~/projects/rssx/deploy/k8s/k8s-rssx-ui-deployment.yaml dest=/tmp'
ansible -i '192.168.50.228,' all -u root -m shell -a 'kubectl delete -f /tmp/k8s-rssx-ui-deployment.yaml'
ansible -i '192.168.50.228,' all -u root -m shell -a 'kubectl create -f /tmp/k8s-rssx-ui-deployment.yaml'
