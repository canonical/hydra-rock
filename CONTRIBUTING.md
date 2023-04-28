# Contributing

## Build and deploy

```bash
rockcraft pack -v
sudo skopeo --insecure-policy copy oci-archive:hydra_2.1.1_amd64.rock docker-daemon:hydra:latest
docker run hydra:latest
```
