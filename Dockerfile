FROM golang:1.22-bookworm AS builder

ARG BUILD_ID=0
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -C cmd/app -ldflags "-X main.BuildId=${BUILD_ID}"

FROM debian:bookworm-slim

COPY --from=builder /src/cmd/app/app /usr/local/bin/app
CMD ["app"]
