# Contributing

## Developing

Please refer to
the [rockcraft](https://canonical-craft-parts.readthedocs-hosted.com/en/latest/reference/index.html)
documentations to learn how to develop a ROCK image.

Please install `pre-commit` hooks to help enforce various validations:

```shell
pre-commit install -t commit-msg
```

## Building and Running Locally

You can build the ROCK image using the following command:

```shell
rockcraft pack -v
```

Assuming the [`skopeo`](https://snapcraft.io/install/skopeo/ubuntu) has been
installed. Import the created ROCK image into Docker:

```shell
sudo /snap/rockcraft/current/bin/skopeo --insecure-policy copy oci-archive:<local-rock-name>.rock docker-daemon:hydra:latest
```

Run a Kratos container using Docker:

```shell
docker run -d \
  --rm \
  --name <container-name> \
  hydra:latest
```
