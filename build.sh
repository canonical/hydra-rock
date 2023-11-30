#!/bin/sh

set -x
# script requires the availability of rockcraft, skopeo, yq and docker in the host system

# export version=$(yq -r '.version' rockcraft.yaml)
rockcraft pack -v

skopeo --insecure-policy copy "oci-archive:hydra_$(yq -r '.version' rockcraft.yaml)_amd64.rock" docker-daemon:$IMAGE

docker push $IMAGE
