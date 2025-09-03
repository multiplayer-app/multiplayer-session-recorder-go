# CLI Example

This example demonstrates how to use the multiplayer OpenTelemetry Go library in a CLI application.

## Setup

1. Set the required environment variable:
   ```bash
   export MULTIPLAYER_OTLP_KEY="your-api-key-here"
   ```

2. Optional environment variables:
   ```bash
   export NODE_ENV=development
   export LOG_LEVEL=debug
   export COMPONENT_NAME=cli-example
   export COMPONENT_VERSION=1.0.0
   export ENVIRONMENT=staging
   export OTLP_TRACES_ENDPOINT=https://api.multiplayer.app/v1/traces
   export OTLP_LOGS_ENDPOINT=https://api.multiplayer.app/v1/logs
   export MULTIPLAYER_OTLP_SPAN_RATIO=0.01
   export MULTIPLAYER_BACKEND_SOURCE=production
   ```

## Running

```bash
cd example/cli
go run .
```

## What it does

1. Initializes OpenTelemetry with the multiplayer session recorder components
2. Creates and starts a debug session
3. Performs some traced work (creates spans)
4. Stops the debug session
5. Waits for telemetry export and exits

## Key Components

- **SessionRecorderIdGenerator**: Generates session-aware trace IDs
- **SessionRecorderHttpTraceExporter**: Exports traces to multiplayer API
- **SessionRecorderHttpLogsExporter**: Exports logs to multiplayer API
- **SessionRecorderSampler**: Custom sampling based on trace ID patterns
- **SessionRecorder**: Manages debug session lifecycle

The example mirrors the functionality of the TypeScript version while using Go idioms and the OpenTelemetry Go SDK.
