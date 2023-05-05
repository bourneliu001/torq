module github.com/lncapital/torq

go 1.18

require (
	github.com/Masterminds/squirrel v1.5.3
	github.com/btcsuite/btcd v0.23.4
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1
	github.com/cockroachdb/errors v1.9.0
	github.com/docker/docker v23.0.3+incompatible
	github.com/docker/go-connections v0.4.0
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/contrib v0.0.0-20201101042839-6a891bf89f19
	github.com/gin-gonic/gin v1.9.0
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/gorilla/websocket v1.5.0
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.0-rc.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.0.0-rc.5
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.3
	github.com/mixer/clock v0.0.0-20210321161542-3ac312e8c7e8
	github.com/pkg/errors v0.9.1
	github.com/playwright-community/playwright-go v0.2000.1
	github.com/prometheus/client_golang v1.14.1-0.20230317203157-0f80f4931424
	github.com/robfig/cron/v3 v3.0.0
	github.com/rs/zerolog v1.27.0
	github.com/signalfx/splunk-otel-go/instrumentation/github.com/jmoiron/sqlx/splunksqlx v1.5.0
	github.com/slack-go/slack v0.12.1
	github.com/ulule/limiter/v3 v3.10.0
	github.com/urfave/cli/v2 v2.8.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.41.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.41.1
	go.opentelemetry.io/otel v1.15.1
	go.opentelemetry.io/otel/exporters/jaeger v1.15.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.15.1
	go.opentelemetry.io/otel/sdk v1.15.1
	go.opentelemetry.io/otel/trace v1.15.1
	go.uber.org/ratelimit v0.2.0
	golang.org/x/exp v0.0.0-20230310171629-522b1b587ee0
	google.golang.org/grpc v1.54.0
	gopkg.in/guregu/null.v4 v4.0.0
	gopkg.in/macaroon.v2 v2.0.0
)

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/andres-erbsen/clock v0.0.0-20160526145045-9e14626cd129 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.2
	github.com/btcsuite/btcd/btcutil v1.1.3
	github.com/bytedance/sonic v1.8.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/containerd/containerd v1.6.19 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/danwakefield/fnmatch v0.0.0-20160403171240-cbb64ac3d964 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/getsentry/sentry-go v0.12.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.11.2 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/goccy/go-json v0.10.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/sessions v1.2.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/iancoleman/strcase v0.2.0
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/juju/loggo v1.0.0 // indirect
	github.com/juju/testing v1.0.1 // indirect
	github.com/klauspost/compress v1.15.6 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.41.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/signalfx/splunk-otel-go v1.4.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql v1.4.0 // indirect
	github.com/signalfx/splunk-otel-go/instrumentation/internal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.9 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opentelemetry.io/otel/metric v0.38.1 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.30.0
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.2.0 // indirect
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v3 v3.0.0
