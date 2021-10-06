test: lint unit-test

export DOCKER_BUILDKIT=1

.PHONY: unit-test
unit-test:
	@docker build . --target unit-test

.PHONY: unit-test-coverage
	@docker build . --target unit-test-coverage
	cat coverage.out

.PHONY: lint
lint:
	@docker build . --target lint
