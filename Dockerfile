FROM golang AS builder
RUN go version

COPY . /src/
WORKDIR /src/
RUN set -x && \
    go get github.com/otiai10/copy && \
    go get gopkg.in/yaml.v2 && \
    go get io/ioutil
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o screeps-launcher ./cmd/screeps-launcher

FROM node:8.11.1-stretch
WORKDIR /screeps
COPY --from=builder /src/screeps-launcher .

ENTRYPOINT ["./screeps-launcher"]
