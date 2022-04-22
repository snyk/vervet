module vervet-underground

go 1.17

require (
	cloud.google.com/go/storage v1.22.0
	github.com/aws/aws-sdk-go-v2 v1.16.2
	github.com/aws/aws-sdk-go-v2/config v1.15.3
	github.com/aws/aws-sdk-go-v2/credentials v1.11.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.5
	github.com/aws/smithy-go v1.11.2
	github.com/frankban/quicktest v1.14.3
	github.com/getkin/kin-openapi v0.94.0
	github.com/gorilla/mux v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/client_model v0.2.0
	github.com/rs/zerolog v1.26.1
	github.com/slok/go-http-metrics v0.10.0
	github.com/snyk/vervet/v4 v4.15.0
	github.com/spf13/viper v1.11.0
	go.uber.org/multierr v1.8.0
	google.golang.org/api v0.75.0
)

require (
	cloud.google.com/go v0.101.0 // indirect
	cloud.google.com/go/compute v1.6.1 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/fsnotify/fsnotify v1.5.3 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.3.0 // indirect
	github.com/googleapis/go-type-adapters v1.0.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.0-beta.8 // indirect
	github.com/prometheus/common v0.34.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/net v0.0.0-20220421235706-1d1ef9303861 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220422154200-b37d22cd5731 // indirect
	google.golang.org/grpc v1.46.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace (
	// Fixes: SNYK-GOLANG-GITHUBCOMGINGONICGIN-1041736
	// From: github.com/slok/go-http-metrics@v0.10.0
	github.com/gin-gonic/gin v1.7.4 => github.com/gin-gonic/gin v1.7.7
	// Fixes: SNYK-GOLANG-GITHUBCOMGOGOPROTOBUFPLUGINUNMARSHAL-1058921
	// From: github.com/spf13/viper@v1.11.0
	github.com/gogo/protobuf v1.1.1 => github.com/gogo/protobuf v1.3.2
	// Fixes: SNYK-GOLANG-GITHUBCOMPROMETHEUSCLIENTGOLANGPROMETHEUSPROMHTTP-2401819
	// From: github.com/spf13/viper@v1.11.0
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v1.11.1

	// Fixes: https://security.snyk.io/vuln/SNYK-GOLANG-GITHUBCOMURFAVENEGRONIV2-1658298
	// https://security.snyk.io/vuln/SNYK-GOLANG-GITHUBCOMURFAVENEGRONI-1658297
	// From: github.com/slok/go-http-metrics@v0.10.0
	github.com/urfave/negroni/v2 v2.0.2 => github.com/urfave/negroni v0.0.0-20211225020424-92731f807096
	// Doesn't work and still flagging failure
 	// github.com/urfave/negroni v1.0.0 => github.com/urfave/negroni v0.0.0-20211225173727-3c3f8059b4bb

	// Fixes: SNYK-GOLANG-GITHUBCOMVALYALAFASTHTTP-2407866
	// From: github.com/slok/go-http-metrics@v0.10.0
	github.com/valyala/fasthttp v1.31.0 => github.com/valyala/fasthttp v1.34.0
)
