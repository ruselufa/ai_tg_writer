module ai_tg_writer

go 1.23.0

toolchain go1.24.4

require (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/gorilla/mux v1.8.1
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.24.3
	// Новые зависимости для мониторинга
	github.com/prometheus/client_golang v1.19.0
	github.com/sirupsen/logrus v1.9.3
	github.com/u2takey/ffmpeg-go v0.5.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.35.0
)

require (
	github.com/aws/aws-sdk-go v1.38.20 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/u2takey/go-utils v0.3.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	go.opentelemetry.io/otel/trace v1.35.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
