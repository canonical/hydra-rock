# Contributing

## Build and deploy

```bash
rockcraft pack -v
sudo skopeo --insecure-policy copy oci-archive:hydra2.1.1_amd64.rock docker-daemon:kratos:latest
docker run kratos:latest
```
