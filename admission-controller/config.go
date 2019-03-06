package admission_controller

import (
	"crypto/tls"

	"github.com/containers-ai/alameda/pkg/framework/datahub"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Config contains the server (the webhook) cert and key.
type Config struct {
	CACertFile        string          `mapstructure:"ca-cert-file"`
	CertFile          string          `mapstructure:"tls-cert-file"`
	KeyFile           string          `mapstructure:"tls-private-key-file"`
	Enable            bool            `mapstructure:"enable"`
	DeployedNamespace string          `mapstructure:"deployed-namespace"`
	Log               *log.Config     `mapstructure:"log"`
	Datahub           *datahub.Config `mapstructure:"datahub"`
}

func NewDefaultConfig() Config {

	defaultDatahubConfig := datahub.NewDefaultConfig()
	defaultLogConfig := log.NewDefaultConfig()

	return Config{
		CACertFile:        "/var/run/secrets/kubernetes.io/serviceaccount/ca.cert",
		CertFile:          "/etc/tls-certs/serverCert.pem",
		KeyFile:           "/etc/tls-certs/serverKey.pem",
		Enable:            false,
		DeployedNamespace: "alameda",
		Log:               &defaultLogConfig,
		Datahub:           &defaultDatahubConfig,
	}
}

func GetK8SClient() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatal(err)
	}
	return clientset
}

func ConfigTLS(config Config, clientset *kubernetes.Clientset) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		glog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}
}
