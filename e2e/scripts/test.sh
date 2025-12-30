#!/bin/bash

set -e

echo "Running e2e tests..."

# Port forward to Prometheus
echo "Setting up port forward to Prometheus..."
kubectl port-forward svc/prometheus 9090:9090 &
PF_PID=$!
trap "kill $PF_PID" EXIT

# Wait for port forward to be ready
sleep 5

# Test 1: Check if Prometheus is scraping metrics
echo "Test 1: Checking if Prometheus is scraping test-client metrics..."
METRICS_RESPONSE=$(curl -s http://localhost:9090/api/v1/targets)
if echo "$METRICS_RESPONSE" | grep -q "active"; then
    echo "✓ Prometheus is active"
else
    echo "✗ Prometheus targets check failed"
    exit 1
fi

# Test 2: Check if rgrpc metrics are present
echo "Test 2: Checking if rgrpc metrics are present..."
# Use a query that will match the metric with any labels
METRICS=$(curl -s 'http://localhost:9090/api/v1/query?query=rgrpc_call_total_ms_count')
if echo "$METRICS" | grep -q "rgrpc_call_total_ms_count"; then
    echo "✓ Found rgrpc_call_total_ms_count metric"
else
    echo "✗ rgrpc_call_total_ms_count metric not found"
    echo "Available rgrpc metrics:"
    curl -s 'http://localhost:9090/api/v1/label/__name__/values' | jq -r '.data[]' | grep rgrpc | head -10 || echo "No rgrpc metrics found"
    echo "Trying alternate query..."
    # Try querying for any rgrpc metric
    ALT_METRICS=$(curl -s 'http://localhost:9090/api/v1/query?query={__name__=~"rgrpc.*"}')
    if echo "$ALT_METRICS" | grep -q "rgrpc"; then
        echo "Found rgrpc metrics with alternate query, but count query failed - this might be a timing issue"
        echo "Waiting 10 seconds for Prometheus to scrape..."
        sleep 10
        METRICS=$(curl -s 'http://localhost:9090/api/v1/query?query=rgrpc_call_total_ms_count')
        if echo "$METRICS" | grep -q "rgrpc_call_total_ms_count"; then
            echo "✓ Found rgrpc_call_total_ms_count metric after wait"
        else
            echo "✗ Still not found after wait"
            exit 1
        fi
    else
        exit 1
    fi
fi

# Test 3: Check if metrics have correct labels
echo "Test 3: Checking if metrics have correct labels (method, remote_ip)..."
METRIC_WITH_LABELS=$(curl -s 'http://localhost:9090/api/v1/query?query=rgrpc_call_total_ms_count')
if echo "$METRIC_WITH_LABELS" | grep -q "method"; then
    echo "✓ Metrics have 'method' label"
else
    echo "✗ Metrics missing 'method' label"
    exit 1
fi

if echo "$METRIC_WITH_LABELS" | grep -q "remote_ip"; then
    echo "✓ Metrics have 'remote_ip' label"
else
    echo "✗ Metrics missing 'remote_ip' label"
    exit 1
fi

# Test 4: Check if histogram buckets exist
echo "Test 4: Checking if histogram buckets exist..."
BUCKETS=$(curl -s 'http://localhost:9090/api/v1/query?query=rgrpc_call_total_ms_bucket')
if echo "$BUCKETS" | grep -q "rgrpc_call_total_ms_bucket"; then
    echo "✓ Found histogram buckets"
else
    echo "✗ Histogram buckets not found"
    exit 1
fi

# Test 5: Check if streaming metrics exist
echo "Test 5: Checking if streaming metrics exist..."
STREAM_METRICS=$(curl -s 'http://localhost:9090/api/v1/query?query=rgrpc_stream_establish_ms_count')
if echo "$STREAM_METRICS" | grep -q "rgrpc_stream_establish_ms_count"; then
    echo "✓ Found streaming metrics"
else
    echo "✗ Streaming metrics not found"
    exit 1
fi

echo ""
echo "All tests passed! ✓"
echo ""
echo "You can view Prometheus at: http://localhost:9090"
echo "Port forward will close when this script exits"

