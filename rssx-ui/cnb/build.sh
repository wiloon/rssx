#!/usr/bin/env bash
# Build the rssx-ui image with Cloud Native Buildpacks.
#
# The single build entry point: homelab Tekton (w10n-config task-cnb-build.yaml)
# runs THIS script inside RSSX_CNB_BUILDER_IMAGE from cnb/images.conf.
#
# The input names keep the ENX_ prefix: they are the env contract of the
# shared w10n-config cnb-build Task, which builds enx-ui the same way.
#
# Required:
#   ENX_APP_DIR       rssx-ui source dir, writable by the CNB user (uid 1000)
#   ENX_IMAGE         image to produce, e.g. registry/rssx-ui:<tag>
# Optional:
#   ENX_CACHE_IMAGE   registry image holding the build cache (none: no cache)
#   ENX_NPM_REGISTRY  npm/pnpm registry for this build (none: public registry)
#   ENX_RUN_IMAGE_MIRROR  host replacing docker.io in the run image reference
#                     (homelab: the in-cluster Nexus Docker Hub proxy)
#   ENX_INSECURE_REGISTRIES  space-separated plain-HTTP registries
#   ENX_PREVIOUS_IMAGE  image to reuse unchanged launch layers from; also
#                     tagged onto the result
#   ENX_REPORT        where to copy the lifecycle report.toml
#   ENX_WORK_DIR      scratch dir for layers/platform (default: /tmp/rssx-cnb)
set -euo pipefail

: "${ENX_APP_DIR:?}" "${ENX_IMAGE:?}"
HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=images.conf
. "$HERE/images.conf"

WORK=${ENX_WORK_DIR:-/tmp/rssx-cnb}
LAYERS=$WORK/layers
PLATFORM=$WORK/platform
BUILDPACKS=$WORK/buildpacks
mkdir -p "$LAYERS" "$PLATFORM/env" "$BUILDPACKS"

if [ -n "${ENX_NPM_REGISTRY:-}" ]; then
  printf '%s' "$ENX_NPM_REGISTRY" > "$PLATFORM/env/npm_config_registry"
  printf '%s' "$ENX_NPM_REGISTRY" > "$PLATFORM/env/pnpm_config_registry"
  echo "build env: npm/pnpm registry = $ENX_NPM_REGISTRY"
fi

# The repo's buildpack, laid out as the lifecycle expects
# (<dir>/<id with / -> _>/<version>/), as the only group in the order.
BP_SRC=$HERE/vite-static
BP_ID=$(sed -n 's/^id = "\(.*\)"/\1/p' "$BP_SRC/buildpack.toml" | head -1)
BP_VERSION=$(sed -n 's/^version = "\(.*\)"/\1/p' "$BP_SRC/buildpack.toml" | head -1)
BP_DIR=$BUILDPACKS/${BP_ID//\//_}/$BP_VERSION
mkdir -p "$BP_DIR"
cp -r "$BP_SRC"/. "$BP_DIR"/
chmod +x "$BP_DIR"/bin/*
printf '[[order]]\n\n  [[order.group]]\n    id = "%s"\n    version = "%s"\n' \
  "$BP_ID" "$BP_VERSION" > "$WORK/order.toml"

RUN_IMAGE=$RSSX_CNB_RUN_IMAGE
if [ -n "${ENX_RUN_IMAGE_MIRROR:-}" ]; then
  RUN_IMAGE=${RUN_IMAGE/#docker.io/$ENX_RUN_IMAGE_MIRROR}
fi

ARGS=(
  -app="$ENX_APP_DIR"
  -layers="$LAYERS"
  -platform="$PLATFORM"
  -buildpacks="$BUILDPACKS"
  -order="$WORK/order.toml"
  -run-image="$RUN_IMAGE"
  -report="$WORK/report.toml"
)
[ -n "${ENX_CACHE_IMAGE:-}" ] && ARGS+=(-cache-image="$ENX_CACHE_IMAGE")
for r in ${ENX_INSECURE_REGISTRIES:-}; do ARGS+=(-insecure-registry="$r"); done
[ -n "${ENX_PREVIOUS_IMAGE:-}" ] && ARGS+=(-previous-image="$ENX_PREVIOUS_IMAGE" -tag="$ENX_PREVIOUS_IMAGE")

echo "builder: $RSSX_CNB_BUILDER_IMAGE"
echo "run image: $RUN_IMAGE"
echo "buildpack: $BP_ID@$BP_VERSION"
# 0.13+ is needed for -insecure-registry.
CNB_PLATFORM_API=0.14 /cnb/lifecycle/creator "${ARGS[@]}" "$ENX_IMAGE"

echo "--- report.toml"
cat "$WORK/report.toml"
[ -n "${ENX_REPORT:-}" ] && cp "$WORK/report.toml" "$ENX_REPORT"
exit 0
