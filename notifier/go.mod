module github.com/containers-ai/alameda/notifier

go 1.13

require (
	github.com/containers-ai/alameda v4.2.293+incompatible
	github.com/containers-ai/api v0.0.0-20190916081203-a47a613d0a1f
	github.com/gogo/protobuf v1.3.0 // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	golang.org/x/net v0.0.0-20190916140828-c8589233b77d // indirect
	google.golang.org/genproto v0.0.0-20190916214212-f660b8655731
	google.golang.org/grpc v1.23.1
	gopkg.in/mail.v2 v2.3.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	k8s.io/apimachinery v0.0.0-20190816221834-a9f1d8a9c101
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0 // indirect
	sigs.k8s.io/controller-runtime v0.2.1
)
