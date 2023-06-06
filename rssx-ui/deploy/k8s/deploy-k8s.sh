#!/bin/sh

cd /home/wiloon/projects/rssx/rssx-ui || exit
yarn install
yarn build
sudo buildah bud -t registry.wiloon.com/rssx-ui:v0.0.1 .
sudo buildah push registry.wiloon.com/rssx-ui:v0.0.1
