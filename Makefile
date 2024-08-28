PORT := 8080
IMAGE := ghcr.io/raeperd/realworld-http-go
TAG := local

all: build test lint docker

build:
	go build -C cmd/app 

test:
	go test -race ./...

lint:
	golangci-lint run

run: build
	./cmd/app/app --port=$(PORT)

docker:
	docker build . -t $(IMAGE):$(TAG)

docker-run: docker 
	docker run --rm -p $(PORT):$(PORT) $(IMAGE):$(TAG)

clean:
	rm cmd/app/app
	docker image rm $(IMAGE):$(TAG)
