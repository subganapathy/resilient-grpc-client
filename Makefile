.PHONY: test bench e2e e2e-clean

# Run unit tests
test:
	go test ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Run e2e tests (requires kind, kubectl, docker)
e2e:
	$(MAKE) -C e2e test

# Clean up e2e resources
e2e-clean:
	$(MAKE) -C e2e cleanup

