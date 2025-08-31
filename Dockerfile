FROM golang:1.25-bookworm AS builder

ARG ARCH=amd64
# Use arm64/32 for other architectures.

WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=${ARCH} \
    go build -o screeps-launcher ./cmd/screeps-launcher

FROM buildpack-deps:buster

ARG UID=1000
ARG GID=1000
RUN <<-EOT bash
    if [[ "${GID}" != "0" ]] ; then
        groupadd --gid ${GID} screeps
    fi
    if [[ "${UID}" != "0" ]] ; then
        useradd --uid ${UID} --gid ${GID} --shell /bin/bash --create-home screeps
    fi
    mkdir /screeps && chown ${UID}:${GID} /screeps
EOT

USER ${UID}:${GID}
VOLUME /screeps
WORKDIR /screeps
COPY --from=builder /app/screeps-launcher /usr/bin/
ENTRYPOINT ["screeps-launcher"]
