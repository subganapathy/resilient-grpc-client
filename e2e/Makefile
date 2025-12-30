.PHONY: setup build deploy test cleanup proto

# Generate proto files
proto:
	cd proto && ./generate.sh

# Set up kind cluster
setup:
	./scripts/setup-kind.sh

# Build and load images
build: proto
	./scripts/build-and-load.sh

# Deploy everything
deploy:
	./scripts/deploy.sh

# Run e2e tests
test:
	./scripts/test.sh

# Clean up resources
cleanup:
	./scripts/cleanup.sh

# Run everything (setup -> build -> deploy -> test)
all: setup build deploy test

