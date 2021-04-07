FROM golang AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o screeps-launcher ./cmd/screeps-launcher

FROM buildpack-deps:buster
RUN groupadd --gid 1000 screeps \
  && useradd --uid 1000 --gid screeps --shell /bin/bash --create-home screeps \
  && mkdir /screeps && chown screeps.screeps /screeps
USER screeps
VOLUME /screeps
WORKDIR /screeps
COPY --from=builder /app/screeps-launcher /usr/bin/
ENTRYPOINT ["screeps-launcher"]
