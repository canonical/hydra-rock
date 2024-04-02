#!/bin/sh

set -x
# script requires the availability of rockcraft, skopeo, yq and docker in the host system
# it also requires sudo permissions to run skopeo

# export version=$(yq -r '.version' rockcraft.yaml)
# rockcraft pack -v

skopeo --insecure-policy copy "oci-archive:identity-platform-login-ui_$(yq -r '.version' rockcraft.yaml)_amd64.rock" docker-daemon:$IMAGE

docker push $IMAGE
