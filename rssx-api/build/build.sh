#!/usr/bin/env bash
# Test, then build the rssx-api image with ko.
#
# The single build entry point: homelab Tekton (w10n-config task-ko-build.yaml)
# runs THIS script inside RSSX_GO_IMAGE from build/build.conf.
#
# The input names keep the ENX_ prefix: they are the env contract of the
# shared w10n-config ko-build Task, which builds enx-api the same way.
#
# Required:
#   ENX_IMAGE         image to produce, registry/repo:tag
# Optional:
#   ENX_PUSH=0        build only, push nothing
#   ENX_BASE_IMAGE_MIRROR  host replacing docker.io in the base image reference
#                     (homelab: the in-cluster Nexus Docker Hub proxy)
#   ENX_INSECURE=1    the registries are plain HTTP (homelab Nexus ports)
#   ENX_IMAGE_REF_FILE  where to write the pushed image as repo@digest
#   RSSX_SKIP_TESTS=1 skip the `go vet` + `go test` gate
#   GOPROXY, GOCACHE, GOMODCACHE  passed through to go / ko as usual
#   Registry credentials: $DOCKER_CONFIG/config.json (ko reads it itself).
set -euo pipefail

: "${ENX_IMAGE:?}"
HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
APP=$(dirname "$HERE")
# shellcheck source=build.conf
. "$HERE/build.conf"

case $ENX_IMAGE in
  */*:*) ;;
  *) echo "ENX_IMAGE must be registry/repo:tag, got $ENX_IMAGE" >&2; exit 1 ;;
esac
REPO=${ENX_IMAGE%:*}
TAG=${ENX_IMAGE##*:}

BASE=$RSSX_BASE_IMAGE
if [ -n "${ENX_BASE_IMAGE_MIRROR:-}" ]; then
  BASE=${BASE/#docker.io/$ENX_BASE_IMAGE_MIRROR}
fi

# Build with the Go in RSSX_GO_IMAGE, never a toolchain go.mod asks to download.
export GOTOOLCHAIN=local
export CGO_ENABLED=0
export KO_DOCKER_REPO=$REPO
export KO_DEFAULTBASEIMAGE=$BASE

echo "go: $(go version)"
echo "ko: $RSSX_KO_VERSION"
echo "base image: $BASE"
echo "image: $ENX_IMAGE"

cd "$APP"

# The gate the Kaniko pipeline ran as a separate go-test Task. Here it shares
# the persistent Go caches with the build.
if [ "${RSSX_SKIP_TESTS:-}" != 1 ]; then
  echo "==> go vet ./..."
  go vet ./...
  echo "==> go test ./..."
  go test ./...
fi

ARGS=(
  --bare
  --tags="$TAG"
  # Image only, no separate SBOM artifact next to it in the registry.
  --sbom=none
)
[ "${ENX_PUSH:-1}" = 0 ] && ARGS+=(--push=false)
[ "${ENX_INSECURE:-}" = 1 ] && ARGS+=(--insecure-registry)

REF=$(go run "github.com/google/ko@$RSSX_KO_VERSION" build "${ARGS[@]}" .)
echo "built: $REF"
[ -n "${ENX_IMAGE_REF_FILE:-}" ] && printf '%s' "$REF" > "$ENX_IMAGE_REF_FILE"
exit 0
