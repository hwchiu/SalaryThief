IMAGE ?= salarythief-collector:dev
GO     ?= go

.PHONY: build test tidy run kind-up kind-down image integration-test test-race

tidy:
	$(GO) mod tidy

build:
	$(GO) build -o bin/collector ./cmd/collector

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

run: build
	./bin/collector -config configs/collector.yaml

image:
	docker build -t $(IMAGE) .

kind-up:
	bash hack/kind-up.sh

kind-down:
	bash hack/kind-down.sh

integration-test:
	bash hack/integration-test.sh
