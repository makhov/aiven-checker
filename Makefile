test-unit:
	go test -mod=vendor ./...
.PHONY: test

test-int:
	go test -mod=vendor ./... -tags=integration
.PHONY: test

test-e2e:
	go test -mod=vendor ./e2e -tags=e2e
.PHONY: test

lint:
	golangci-lint run --verbose
.PHONY: lint

build-image:
	docker build -t aiven-checker
.PHONY: build-image


