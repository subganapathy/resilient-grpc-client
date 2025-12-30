# End-to-End Testing

This directory contains end-to-end tests that validate the resilient-grpc-client library in a real Kubernetes environment using kind (Kubernetes in Docker).

## Prerequisites

- [kind](https://kind.sigs.k8s.io/) installed
- [kubectl](https://kubernetes.io/docs/tasks/tools/) installed
- [docker](https://docs.docker.com/get-docker/) installed
- Go 1.25+
- [protoc](https://grpc.io/docs/protoc-installation/) installed
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins (installed via `go install`)

## Test Structure

```
e2e/
├── test-server/          # Simple gRPC test server
├── test-client/          # Test client using rgrpc library
├── manifests/            # Kubernetes manifests
│   ├── prometheus/       # Prometheus deployment
│   └── apps/             # Test server and client deployments
├── scripts/              # Helper scripts
└── README.md            # This file
```

## Running the Tests

**Quick start (from repository root):**
```bash
make e2e  # Runs the full e2e test suite
```

**Or from the e2e directory:**
```bash
cd e2e
make all  # Runs: setup -> build -> deploy -> test
```

**Step by step (from e2e directory):**

1. **Generate proto files:**
   ```bash
   cd e2e
   make proto
   # Or manually: cd proto && ./generate.sh
   ```

2. **Start the kind cluster:**
   ```bash
   ./scripts/setup-kind.sh
   # Or: make setup
   ```

3. **Build and load images:**
   ```bash
   ./scripts/build-and-load.sh
   # Or: make build
   ```

4. **Deploy everything:**
   ```bash
   ./scripts/deploy.sh
   # Or: make deploy
   ```

5. **Run the e2e test:**
   ```bash
   ./scripts/test.sh
   # Or: make test
   ```

6. **Clean up:**
   ```bash
   ./scripts/cleanup.sh
   # Or: make cleanup
   ```

## What Gets Tested

- Metrics are emitted to Prometheus
- Histogram metrics are correctly formatted
- Labels (method, remote_ip) are present
- Metrics can be queried via Prometheus API
- Both unary and streaming RPCs emit metrics

