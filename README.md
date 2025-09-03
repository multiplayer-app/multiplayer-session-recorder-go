![Description](./docs/img/header-go.png)

<div align="center">
<a href="https://github.com/multiplayer-app/multiplayer-session-recorder-javascript">
  <img src="https://img.shields.io/github/stars/multiplayer-app/multiplayer-session-recorder-javascript?style=social&label=Star&maxAge=2592000" alt="GitHub stars">
</a>
  <a href="https://github.com/multiplayer-app/multiplayer-session-recorder-javascript/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/multiplayer-app/multiplayer-session-recorder-javascript" alt="License">
  </a>
  <a href="https://multiplayer.app">
    <img src="https://img.shields.io/badge/Visit-multiplayer.app-blue" alt="Visit Multiplayer">
  </a>
  
</div>
<div>
  <p align="center">
    <a href="https://x.com/trymultiplayer">
      <img src="https://img.shields.io/badge/Follow%20on%20X-000000?style=for-the-badge&logo=x&logoColor=white" alt="Follow on X" />
    </a>
    <a href="https://www.linkedin.com/company/multiplayer-app/">
      <img src="https://img.shields.io/badge/Follow%20on%20LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white" alt="Follow on LinkedIn" />
    </a>
    <a href="https://discord.com/invite/q9K3mDzfrx">
      <img src="https://img.shields.io/badge/Join%20our%20Discord-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Join our Discord" />
    </a>
  </p>
</div>

# Multiplayer Full Stack Session Recorder

The Multiplayer Full Stack Session Recorder is a powerful tool that offers deep session replays with insights spanning frontend screens, platform traces, metrics, and logs. It helps your team pinpoint and resolve bugs faster by providing a complete picture of your backend system architecture. No more wasted hours combing through APM data; the Multiplayer Full Stack Session Recorder does it all in one place.

## Install

```bash
go get github.com/multiplayer-app/multiplayer-otlp-go
```

## Set up backend services

### Route traces and logs to Multiplayer

Multiplayer Full Stack Session Recorder is built on top of OpenTelemetry.

### New to OpenTelemetry?

No problem. You can set it up in a few minutes. If your services don't already use OpenTelemetry, you'll first need to install the OpenTelemetry libraries. Detailed instructions for this can be found in the [OpenTelemetry documentation](https://opentelemetry.io/docs/).

### Already using OpenTelemetry?

You have two primary options for routing your data to Multiplayer:

***Direct Exporter***: This option involves using the Multiplayer Exporter directly within your services. It's a great choice for new applications or startups because it's simple to set up and doesn't require any additional infrastructure. You can configure it to send all session recording data to Multiplayer while optionally sending a sampled subset of data to your existing observability platform.

***OpenTelemetry Collector***: For large, scaled platforms, we recommend using an OpenTelemetry Collector. This approach provides more flexibility by having your services send all telemetry to the collector, which then routes specific session recording data to Multiplayer and other data to your existing observability tools.


### Option 1: Direct Exporter

Send OpenTelemetry data from your services to Multiplayer and optionally other destinations (e.g., OpenTelemetry Collectors, observability platforms, etc.).

This is the quickest way to get started, but consider using an OpenTelemetry Collector (see [Option 2](#option-2-opentelemetry-collector) below) if you're scalling or a have a large platform.

```go
import (
    "github.com/multiplayer-app/multiplayer-otlp-go/exporters"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
)

// set up Multiplayer exporters. Note: GRPC exporters are also available.
// see: NewSessionRecorderGrpcTraceExporter and NewSessionRecorderGrpcLogsExporter
multiplayerTraceExporter, err := exporters.NewSessionRecorderHttpTraceExporter(
    "MULTIPLAYER_OTLP_KEY", // note: replace with your Multiplayer OTLP key
)
if err != nil {
    log.Fatal(err)
}

multiplayerLogExporter, err := exporters.NewSessionRecorderHttpLogsExporter(
    "MULTIPLAYER_OTLP_KEY", // note: replace with your Multiplayer OTLP key
)
if err != nil {
    log.Fatal(err)
}

// Multiplayer exporter wrappers filter out session recording atrtributes before passing to provided exporter
standardTraceExporter, _ := otlptracehttp.New(ctx)
traceExporter := exporters.NewSessionRecorderTraceExporterWrapper(standardTraceExporter)

standardLogExporter, _ := otlploghttp.New(ctx)
logExporter := exporters.NewSessionRecorderLogsExporterWrapper(standardLogExporter)
```

### Option 2: OpenTelemetry Collector

If you're scalling or a have a large platform, consider running a dedicated collector. See the Multiplayer OpenTelemetry collector [repository](https://github.com/multiplayer-app/multiplayer-otlp-collector) which shows how to configure the standard OpenTelemetry Collector to send data to Multiplayer and optional other destinations.

Add standard [OpenTelemetry code](https://opentelemetry.io/docs/languages/go/exporters/) to export OTLP data to your collector.

See a basic example below:

```go
import (
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
)

traceExporter, err := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint("http://<OTLP_COLLECTOR_URL>/v1/traces"),
    otlptracehttp.WithHeaders(map[string]string{
        // ...
    }),
)

logExporter, err := otlploghttp.New(ctx,
    otlploghttp.WithEndpoint("http://<OTLP_COLLECTOR_URL>/v1/logs"),
    otlploghttp.WithHeaders(map[string]string{
        // ...
    }),
)
```

### Capturing request/response and header content

In addition to sending traces and logs, you need to capture request and response content. We offer two solutions for this:

***In-Service Code Capture:*** You can use our libraries to capture, serialize, and mask request/response and header content directly within your service code. This is an easy way to get started, especially for new projects, as it requires no extra components in your platform.

***Multiplayer Proxy:*** Alternatively, you can run a [Multiplayer Proxy](https://github.com/multiplayer-app/multiplayer-proxy) to handle this outside of your services. This is ideal for large-scale applications and supports all languages, including those like Java that don't allow for in-service request/response hooks. The proxy can be deployed in various ways, such as an Ingress Proxy, a Sidecar Proxy, or an Embedded Proxy, to best fit your architecture.

### Option 1: In-Service Code Capture

The Multiplayer Session Recorder library provides utilities for capturing request, response and header content. See example below:

```go
import (
    "github.com/multiplayer-app/multiplayer-otlp-go/middleware"
    "net/http"
)

// configure middleware options
options := middleware.NewMiddlewareOptions(
    middleware.WithCaptureHeaders(true),
    middleware.WithCaptureBody(true),
    middleware.WithMaskHeadersEnabled(true),
    middleware.WithMaskBodyEnabled(false),
    // set the maximum request/response content size (in bytes) that will be captured
    // any request/response content greater than size will be not included in session recordings
    middleware.WithMaxPayloadSizeBytes(500000),
    // list of headers to mask in request/response headers
    middleware.WithMaskHeadersList([]string{"set-cookie", "Authorization", "cookie"}),
)

// Apply middleware to your HTTP handlers
handler := middleware.WithRequestData(http.HandlerFunc(yourHandler), options)
handler = middleware.WithResponseData(handler, options)

```

### Option 2: Multiplayer Proxy

The Multiplayer Proxy enables capturing request/response and header content without changing service code. See instructions at the [Multiplayer Proxy repository](https://github.com/multiplayer-app/multiplayer-proxy).

## Set up CLI app

The Multiplayer Full Stack Session Recorder can be used inside the CLI apps.

The [Multiplayer Time Travel Demo](https://github.com/multiplayer-app/multiplayer-time-travel-platform) includes an example [node.js CLI app](https://github.com/multiplayer-app/multiplayer-time-travel-platform/tree/main/clients/go-cli-app).

See an additional example below.

### Quick start

Use the following code below to initialize and run the session recorder.

Example for Session Recorder initialization relies on [opentelemetry.go](./example/cli/opentelemetry.go) file. Copy that file and put next to quick start code.

```go
// IMPORTANT: set up OpenTelemetry
// for an example see ./example/cli/opentelemetry.go
import (
    "github.com/multiplayer-app/multiplayer-otlp-go/session_recorder"
    "github.com/multiplayer-app/multiplayer-otlp-go/types"
)

// initialize session recorder
sr := session_recorder.NewSessionRecorder()

config := session_recorder.SessionRecorderConfig{
    APIKey:           "MULTIPLAYER_OTLP_KEY", // note: replace with your Multiplayer OTLP key
    TraceIDGenerator: idGenerator, // from OpenTelemetry setup
    ResourceAttributes: map[string]interface{}{
        "serviceName":  "{YOUR_APPLICATION_NAME}",
        "version":      "{YOUR_APPLICATION_VERSION}",
        "environment":  "{YOUR_APPLICATION_ENVIRONMENT}",
    },
}

err := sr.Init(config)
if err != nil {
    log.Fatal(err)
}

session := &session_recorder.Session{
    Name: "This is test session",
    SessionAttributes: map[string]interface{}{
        "accountId":   "687e2c0d3ec8ef6053e9dc97",
        "accountName": "Acme Corporation",
    },
}

err = sr.Start(types.SESSION_TYPE_PLAIN, session)
if err != nil {
    log.Fatal(err)
}

// do something here

err = sr.Stop(nil)
if err != nil {
    log.Fatal(err)
}
```

Replace the placeholders with your application’s version, name, environment, and API key.

## License

MIT — see [LICENSE](./LICENSE).
