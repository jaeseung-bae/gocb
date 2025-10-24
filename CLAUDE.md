# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the official Couchbase Go SDK (gocb v2), a pure Go client library for connecting to Couchbase clusters. It provides high-level APIs for key-value operations, queries (N1QL), analytics, full-text search, and transactions.

## Testing

### Running Tests

```bash
# Run all tests (unit + integration) with race detection
go test -race ./...

# Run only unit tests (skip integration tests)
go test -short ./

# Run with coverage
go test -coverprofile=cover.out ./

# Fast test (alias for short mode)
make fasttest
```

### Integration Tests

By default, integration tests run against a mock Couchbase Server (gocaves). To run against a real server:

```bash
# Using environment variables
export GOCBSERVER="couchbase://localhost"
export GOCBUSER="Administrator"
export GOCBPASS="password"
export GOCBBUCKET="default"
export GOCBVER="7.6.0"
go test ./...

# Or using command-line flags
go test -server="couchbase://localhost" -user="Administrator" -pass="password" -bucket="default" -version="7.6.0" ./...
```

**Important test flags:**
- `-short`: Skip integration tests
- `-disable-logger=true`: Disable verbose logging (useful for benchmarks)
- `-collection-name=<name>`: Test against specific collection
- `-scope-name=<name>`: Test against specific scope
- `-features=<flags>`: Enable/disable features (e.g., `+xattrs,-transactions`)
- `-certs-path=<path>`: Path to TLS certificates directory

### Running Specific Tests

```bash
# Run a single test
go test -run TestBasicOps ./

# Run tests matching a pattern
go test -run "TestBucket.*" ./

# Run benchmarks
make bench
# or
go test -bench=. -run=none --disable-logger=true
```

## Development Commands

### Setup Development Environment

```bash
make devsetup
```

This installs required tools:
- `golangci-lint` v1.61.0 (linting)
- `mockery` v2.46.2 (mock generation)

### Linting

```bash
make lint
```

### Pre-commit Check

```bash
# Runs lint, short tests, coverage, and race detection
make check
```

### Updating Mocks

When provider interfaces change, regenerate mocks:

```bash
make updatemocks
```

### Updating Test Cases

The project uses git submodules for shared test cases:

```bash
make updatetestcases
```

## Code Architecture

### Dual Backend Support

The SDK supports two backend implementations:

1. **gocbcore** (default): Binary protocol using `github.com/couchbase/gocbcore/v10`
   - Files: `*_core.go`, `client_core.go`
   - Connection manager: `stdConnectionMgr`

2. **Protostellar**: gRPC protocol using `github.com/couchbase/gocbcoreps`
   - Files: `*_ps.go`, `client_ps.go`
   - Connection manager: `psConnectionMgr`

The backend is selected automatically based on the connection string scheme.

### Provider Pattern

The SDK uses a provider abstraction pattern to support multiple backends:

- Provider interfaces are defined for each service (e.g., `kvProvider`, `queryProvider`, `analyticsProvider`)
- Implementations exist for both backends (`*_core.go` and `*_ps.go`)
- The `connectionManager` interface abstracts connection lifecycle
- `providerController[T]` manages provider lifecycle and retries

Key provider types:
- `kvProvider`: Key-value operations
- `queryProvider`: N1QL queries
- `analyticsProvider`: Analytics queries
- `searchProvider`: Full-text search
- `viewProvider`: Views (legacy)
- `diagnosticsProvider`: Health checks and diagnostics
- `transactionsProvider`: Distributed transactions

### Object Hierarchy

```
Cluster
  └── Bucket
       └── Scope
            └── Collection (main API for KV operations)
```

- **Cluster**: Entry point, manages connections, executes cluster-level operations (queries, analytics, user management)
- **Bucket**: Represents a bucket, manages collections/scopes, executes bucket-level operations (views)
- **Scope**: Logical grouping of collections, supports scope-level queries
- **Collection**: Primary interface for key-value operations (Get, Insert, Upsert, Remove, etc.)

### Key Components

- **Transcoder**: Serialization/deserialization of documents (JSON, raw, legacy)
- **RetryStrategy**: Configurable retry logic for operations
- **Tracer/Meter**: OpenTelemetry integration for observability
- **Transactions**: Distributed ACID transactions support
- **CircuitBreaker**: Failure detection and recovery

### Async Operation Management

The SDK internally uses callbacks (gocbcore style) but exposes synchronous APIs. The `asyncOpManager` handles this conversion:

```go
opm := newAsyncOpManager(ctx)
err := opm.Wait(provider.SomeOperation(opts, func(res *Result, err error) {
    // callback handling
    opm.Resolve() // or opm.Reject()
}))
```

### Subdirectories

- `search/`: Full-text search query builders and facets
- `vector/`: Vector search support
- `testdata/`: Test fixtures and data files

## Common Patterns

### Error Handling

Errors follow specific hierarchies:
- `KeyValueError`: KV operation errors (DocumentNotFound, CasMismatch, etc.)
- `QueryError`, `AnalyticsError`, `SearchError`, `ViewError`: Service-specific errors
- `TimeoutError`: Operation timeouts
- Generic errors: Authentication, network, etc.

### Context Usage

Operations support `context.Context` for cancellation and timeouts through options structs (e.g., `GetOptions.Context`).

### Timeouts Configuration

Timeouts are hierarchical:
1. Per-operation timeout (in options)
2. Service-level timeout (in `TimeoutsConfig`)
3. Default timeout (2.5s for KV, 75s for queries)

### Memory Considerations

**gocbcore DCP buffers**: If using DCP (Data Change Protocol), be aware that v7.1.18+ uses large buffered channels (default 8MB = ~350k items per connection). Configure via:
```go
config.DcpBufferSize = 1 * 1024 * 1024 // Reduce to 1MB
```

## Issue Tracking

- Issues are tracked on [issues.couchbase.com](http://www.couchbase.com/issues/browse/GOCBC)
- Prefix: GOCBC-XXXX
- All significant changes should reference a JIRA ticket

## Documentation

- API Reference: [pkg.go.dev/github.com/couchbase/gocb](https://pkg.go.dev/github.com/couchbase/gocb)
- Official Docs: [docs.couchbase.com/go-sdk](https://docs.couchbase.com/go-sdk/current/hello-world/overview.html)
- Community: [Discord](https://discord.com/invite/sQ5qbPZuTh) and [Forums](https://forums.couchbase.com/c/go-sdk/23)
