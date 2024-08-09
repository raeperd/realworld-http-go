FROM golang:1.22-bullseye as builder
WORKDIR /src
COPY . /src
RUN make build CGO_ENABLED=0 

FROM scratch
COPY --from=builder /src/cmd/app/app /bin/app
ENTRYPOINT ["/bin/app"]
CMD ["--port=8080"]
