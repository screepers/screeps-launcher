FROM golang AS builder
COPY . /go/src/github/ags131/screeps-launcher
WORKDIR /go/src/github/ags131/screeps-launcher
RUN go get ./...
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -o screeps-launcher ./cmd/screeps-launcher

FROM node:8.15
VOLUME /screeps
WORKDIR /screeps
COPY --from=builder /go/src/github/ags131/screeps-launcher .

ENTRYPOINT ["./screeps-launcher"]
