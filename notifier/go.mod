module github.com/containers-ai/alameda/notifier

go 1.12

require (
	github.com/containers-ai/alameda v4.1.113+incompatible
	github.com/containers-ai/api v0.0.0-20190729042350-e83e4c249904
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.6.0
	github.com/onsi/gomega v1.4.2
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0-beta.4
)
