module github.com/letsencrypt/boulder

go 1.21

require (
	github.com/aws/aws-sdk-go-v2 v1.24.1
	github.com/aws/aws-sdk-go-v2/config v1.26.3
	github.com/aws/aws-sdk-go-v2/service/s3 v1.48.0
	github.com/aws/smithy-go v1.19.0
	github.com/eggsampler/acme/v3 v3.4.0
	github.com/go-logr/stdr v1.2.2
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/google/certificate-transparency-go v1.1.6
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/jmhodges/clock v1.2.0
	github.com/letsencrypt/borp v0.0.0-20230707160741-6cc6ce580243
	github.com/letsencrypt/challtestsrv v1.2.1
	github.com/letsencrypt/pkcs11key/v4 v4.0.0
	github.com/letsencrypt/validator/v10 v10.0.0-20230215210743-a0c7dfc17158
	github.com/miekg/dns v1.1.55
	github.com/miekg/pkcs11 v1.1.1
	github.com/nxadm/tail v1.4.11
	github.com/prometheus/client_golang v1.15.1
	github.com/prometheus/client_model v0.4.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/titanous/rocacheck v0.0.0-20171023193734-afe73141d399
	github.com/weppos/publicsuffix-go v0.30.2-0.20230730094716-a20f9abcc222
	github.com/zmap/zcrypto v0.0.0-20231219022726-a1f61fb1661c
	github.com/zmap/zlint/v3 v3.6.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.46.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.21.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
	golang.org/x/crypto v0.18.0
	golang.org/x/net v0.20.0
	golang.org/x/sync v0.4.0
	golang.org/x/term v0.16.0
	golang.org/x/text v0.14.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	gopkg.in/go-jose/go-jose.v2 v2.6.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.14 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.7 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.16.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/poy/onpar v1.1.2 // indirect
	github.com/prometheus/common v0.42.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/proto/otlp v1.0.0 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/tools v0.8.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231002182017-d307bd883b97 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231212172506-995d672761c0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
)

// Versions of go-sql-driver/mysql >1.5.0 introduce performance regressions for
// us, so we exclude them.

// This version is required by parts of the honeycombio/beeline-go package
exclude github.com/go-sql-driver/mysql v1.6.0

// This version is required by borp
exclude github.com/go-sql-driver/mysql v1.7.1
