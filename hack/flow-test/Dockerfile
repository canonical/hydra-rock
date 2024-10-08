FROM --platform=$BUILDPLATFORM golang:1.21-bullseye AS builder

LABEL org.opencontainers.image.source=https://github.com/canonical/hydra-rock/hack/flow-test

ARG SKAFFOLD_GO_GCFLAGS
ARG TARGETOS
ARG TARGETARCH

ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /var/app

COPY . .

RUN go build -o app ./

# FROM gcr.io/distroless/static:nonroot
FROM gcr.io/distroless/base


LABEL org.opencontainers.image.source=https://github.com/canonical/hydra-rock/hack/flow-test

WORKDIR /

COPY --from=builder /var/app/app /app

CMD ["/app"]
