IMAGE ?= salarythief-collector:dev
GO     ?= go

.PHONY: build test tidy run kind-up kind-down image

tidy:
	$(GO) mod tidy

build:
	$(GO) build -o bin/collector ./cmd/collector

test:
	$(GO) test ./...

run: build
	./bin/collector -config configs/collector.yaml

image:
	docker build -t $(IMAGE) .

kind-up:
	bash hack/kind-up.sh

kind-down:
	bash hack/kind-down.sh
