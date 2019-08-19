module github.com/containers-ai/alameda/notifier

go 1.12

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/blang/semver v3.5.0+incompatible // indirect
	github.com/containers-ai/alameda v4.2.257+incompatible
	github.com/containers-ai/api v0.0.0-20190814025936-612c4c93ff8b
	github.com/coreos/etcd v3.3.13+incompatible // indirect
	github.com/coreos/go-oidc v0.0.0-20180117170138-065b426bd416 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/docker/spdystream v0.0.0-20160310174837-449fdfce4d96 // indirect
	github.com/elazarl/goproxy v0.0.0-20170405201442-c4fc26588b6e // indirect
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/validate v0.19.2 // indirect
	github.com/google/go-cmp v0.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	github.com/pquerna/cachecontrol v0.0.0-20171018203845-0dec1b30a021 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/streadway/amqp v0.0.0-20190815230801-eade30b20f1d
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a // indirect
	google.golang.org/appengine v1.5.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.2.2 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.0.0-20190814140513-6483f25b1faf // indirect
	k8s.io/utils v0.0.0-20190809000727-6c36bc71fc4a // indirect
	modernc.org/cc v1.0.0
	modernc.org/golex v1.0.0
	modernc.org/mathutil v1.0.0
	modernc.org/strutil v1.0.0
	modernc.org/xc v1.0.0
	sigs.k8s.io/controller-runtime v0.2.0-rc.0
	sigs.k8s.io/structured-merge-diff v0.0.0-20190724202554-0c1d754dd648 // indirect
)

replace (
	modernc.org/cc v1.0.0 => gitlab.com/cznic/cc v1.0.0
	modernc.org/golex v1.0.0 => gitlab.com/cznic/golex v1.0.0
	modernc.org/mathutil v1.0.0 => gitlab.com/cznic/mathutil v1.0.0
	modernc.org/strutil v1.0.0 => gitlab.com/cznic/strutil v1.0.0
	modernc.org/xc v1.0.0 => gitlab.com/cznic/xc v1.0.0
)
