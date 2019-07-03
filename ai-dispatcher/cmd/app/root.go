package app

import (
	"os"
	"strings"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/dispatcher"
	alameda_app "github.com/containers-ai/alameda/cmd/app"
	"github.com/containers-ai/alameda/datahub"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

var (
	cfgFile             string
	logRotateOutputFile string

	scope  *log.Scope
	config datahub.Config
)

const (
	envVarPrefix = "ALAMEDA_AI_DISPATCHER"

	defaultRotationMaxSizeMegabytes = 100
	defaultRotationMaxBackups       = 7
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(alameda_app.VersionCmd)
	rootCmd.AddCommand(ProbeCmd)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/alameda/ai-dispatcher/ai-dispatcher.yml", "The path to ai-dispatcher configuration file.")
	rootCmd.PersistentFlags().StringVar(&logRotateOutputFile, "log-output-file", "/var/log/alameda/alameda-ai-dispatcher.log", "The path of log file.")
}

var rootCmd = &cobra.Command{
	Use:   "ai-dispatcher",
	Short: "AI dispatcher sends predicted jobs to queue",
	Long: `AI dispatcher send predicted jobs to queue
			including nodes and pods`,
	Run: func(cmd *cobra.Command, args []string) {
		initLogger()
		setLoggerScopesWithConfig()
		datahubAddr := viper.GetString("datahub_address")
		if datahubAddr == "" {
			scope.Errorf("No configuration of datahub address.")
			return
		}
		conn, err := grpc.Dial(datahubAddr, grpc.WithInsecure())
		if err != nil {
			return
			scope.Errorf("Datahub connection constructs failed. %s", err.Error())
		}

		queueUrl := viper.GetString("queue.url")
		if queueUrl == "" {
			scope.Errorf("No configuration of queue url.")
			return
		}
		queueConn, err := amqp.Dial(queueUrl)
		if err != nil {
			scope.Errorf("Queue connection constructs failed. %s", err.Error())
			return
		}
		defer queueConn.Close()
		defer conn.Close()
		dp := dispatcher.NewDispatcher(conn, queueConn)
		dp.Start()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		scope.Errorf("%s", err.Error())
		os.Exit(1)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.SetEnvPrefix(envVarPrefix)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	if err := viper.ReadInConfig(); err != nil {
		scope.Errorf("Can't read config: %s", err.Error())
		os.Exit(1)
	}
}

func initLogger() {

	opt := log.DefaultOptions()
	opt.RotationMaxSize = defaultRotationMaxSizeMegabytes
	opt.RotationMaxBackups = defaultRotationMaxBackups
	opt.RotateOutputPath = logRotateOutputFile

	err := log.Configure(opt)
	if err != nil {
		panic(err)
	}

	scope = log.RegisterScope("app", "ai-dispatcher app", 0)
}

func setLoggerScopesWithConfig() {
	setLogCallers := true
	if viper.IsSet("log.set_logcallers") {
		setLogCallers = viper.GetBool("log.set_logcallers")
	}

	outputLevel := "none"
	if viper.IsSet("log.output_level") {
		outputLevel = viper.GetString("log.output_level")
	}

	stackTraceLvl := "none"
	if viper.IsSet("log.stacktrace_level") {
		stackTraceLvl = viper.GetString("log.stacktrace_level")
	}
	for _, scope := range log.Scopes() {
		scope.SetLogCallers(setLogCallers)
		if outputLvl, ok := log.StringToLevel(outputLevel); ok {
			scope.SetOutputLevel(outputLvl)
		}
		if stacktraceLevel, ok := log.StringToLevel(stackTraceLvl); ok {
			scope.SetStackTraceLevel(stacktraceLevel)
		}
	}
}
