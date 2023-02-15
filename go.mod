module github.com/sysdiglabs/stackdriver-webhook-bridge

go 1.13

require (
	cloud.google.com/go/logging v1.0.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/golang/protobuf v1.4.3
	github.com/prometheus/client_golang v1.11.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.4.0
	google.golang.org/api v0.13.0
	google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/apiserver v0.17.0
)
