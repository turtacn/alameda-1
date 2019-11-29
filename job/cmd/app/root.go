package app

import (
	"errors"
	"strings"

	k8SUtils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const envVarPrefix = "ALAMEDA_JOB"

var configFile string
var scope = logUtil.RegisterScope("job", "job", 0)
var k8sCli client.Client

func initK8SClient() {
	// Instance kubernetes client
	k8sClient, err := k8SUtils.NewK8SClient()
	if err != nil {
		panic(errors.New("Get kubernetes client failed: " + err.Error()))
	} else {
		k8sCli = k8sClient
	}
}

func initConfig() {
	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.SetConfigFile(configFile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(errors.New("Read configuration failed: " + err.Error()))
	}
}

var RootCmd = &cobra.Command{
	Use:   "job",
	Short: "alameda job",
	Long:  "",
}

func init() {
	RootCmd.AddCommand(InstallCmd)
	RootCmd.AddCommand(UninstallCmd)

	RootCmd.PersistentFlags().StringVar(&configFile,
		"config", "/etc/alameda/job/job.toml",
		"The path to job configuration file.")
}
