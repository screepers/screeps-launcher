FROM --platform=$BUILDPLATFORM golang:1.26-trixie AS builder

# TARGETARCH and TARGETVARIANT are provided by buildx
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT#v} \
    go build -o screeps-launcher ./cmd/screeps-launcher

FROM buildpack-deps:trixie

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
