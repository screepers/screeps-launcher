FROM golang:1.23-bookworm AS builder

WORKDIR /app

ARG ARCH=amd64
# Use arm64/32 for other architectures.
COPY screeps-launcher .
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=${ARCH} \
    go build -o screeps-launcher ./cmd/screeps-launcher


FROM buildpack-deps:buster

COPY --from=builder /app/screeps-launcher /usr/bin/

RUN groupadd --gid 1000 screeps \
  && useradd --uid 1000 --gid screeps --shell /bin/bash --create-home screeps \
  && mkdir /screeps && chown screeps.screeps /screeps

USER screeps
VOLUME /screeps
WORKDIR /screeps

EXPOSE 21025/tcp
ENTRYPOINT ["screeps-launcher"]
