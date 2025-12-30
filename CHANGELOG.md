# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-01-XX

### Added
- Initial release
- Automatic latency breakdown metrics (stream_establish_ms, send_stall_ms, response_wait_ms)
- TCP-level diagnostics (RTT, congestion window, retransmissions) for Linux
- Support for both unary and streaming gRPC calls
- OpenTelemetry integration with Prometheus export
- Configurable client-side load balancing for Kubernetes headless services
- Background TCP sampling to avoid hot-path syscalls

