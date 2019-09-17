package app

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/containers-ai/alameda/ai-dispatcher/pkg/dispatcher"
	"github.com/containers-ai/alameda/ai-dispatcher/pkg/queue"
	alameda_app "github.com/containers-ai/alameda/cmd/app"
	"github.com/containers-ai/alameda/pkg/utils/log"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	cfgFile             string
	logRotateOutputFile string

	scope *log.Scope
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config",
		"/etc/alameda/ai-dispatcher/ai-dispatcher.toml",
		"The path to ai-dispatcher configuration file.")
	rootCmd.PersistentFlags().StringVar(&logRotateOutputFile, "log-output-file",
		"/var/log/alameda/alameda-ai-dispatcher.log", "The path of log file.")
}

var rootCmd = &cobra.Command{
	Use:   "ai-dispatcher",
	Short: "AI dispatcher sends predicted jobs to queue",
	Long: `AI dispatcher send predicted jobs to queue
			including nodes and pods`,
	Run: func(cmd *cobra.Command, args []string) {
		initLogger()
		setLoggerScopesWithConfig()
		datahubAddr := viper.GetString("datahub.address")
		if datahubAddr == "" {
			scope.Errorf("No configuration of datahub address.")
			return
		}
		datahubConnRetry := viper.GetInt("datahub.connRetry")
		conn, err := grpc.Dial(datahubAddr, grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
				grpc_retry.WithMax(uint(datahubConnRetry)))))
		if err != nil {
			scope.Errorf("Datahub connection constructs failed. %s", err.Error())
			return
		}

		queueURL := viper.GetString("queue.url")
		if queueURL == "" {
			scope.Errorf("No configuration of queue url.")
			return
		}

		defer conn.Close()
		granularities := viper.GetStringSlice("serviceSetting.granularities")
		predictUnits := viper.GetStringSlice("serviceSetting.predictUnits")
		modelMapper := dispatcher.NewModelMapper(predictUnits, granularities)
		if viper.GetBool("model.enabled") {
			go modelCompleteNotification(modelMapper)
		}
		dp := dispatcher.NewDispatcher(conn, granularities, predictUnits, modelMapper)
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
		panic(err)
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
	if viper.IsSet("log.setLogcallers") {
		setLogCallers = viper.GetBool("log.setLogcallers")
	}

	outputLevel := "none"
	if viper.IsSet("log.outputLevel") {
		outputLevel = viper.GetString("log.outputLevel")
	}

	stackTraceLvl := "none"
	if viper.IsSet("log.stacktraceLevel") {
		stackTraceLvl = viper.GetString("log.stacktraceLevel")
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

func modelCompleteNotification(modelMapper *dispatcher.ModelMapper) {
	reconnectInterval := viper.GetInt64("queue.consumer.reconnectInterval")
	queueConnRetryItvMS := viper.GetInt64("queue.retry.connectIntervalMs")
	modelCompleteQueue := "model_complete"
	queueURL := viper.GetString("queue.url")
	for {
		queueConn := queue.GetQueueConn(queueURL, queueConnRetryItvMS)
		queueConsumer := queue.NewRabbitMQConsumer(queueConn)
		for {
			msg, ok, err := queueConsumer.ReceiveJsonString(modelCompleteQueue)
			if err != nil {
				scope.Errorf("Get message from model complete queue error: %s", err.Error())
				break
			}
			if !ok {
				scope.Infof("No jobs found in queue %s, retry to get jobs next %v seconds",
					modelCompleteQueue, reconnectInterval)
				time.Sleep(time.Duration(reconnectInterval) * time.Second)
				break
			}

			var msgMap map[string]interface{}
			msgByte := []byte(msg)
			if err := json.Unmarshal(msgByte, &msgMap); err != nil {
				scope.Errorf("decode model complete job from queue failed: %s", err.Error())
				continue
			}

			unit := msgMap["unit"].(map[string]interface{})
			unitType := msgMap["unit_type"].(string)
			dataGranularity := msgMap["data_granularity"].(string)
			if unitType == dispatcher.UnitTypeNode {
				nodeName := unit["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity, nodeName)
			} else if unitType == dispatcher.UnitTypePod {
				podNamespacedName := unit["namespaced_name"].(map[string]interface{})
				podNS := podNamespacedName["namespace"].(string)
				podName := podNamespacedName["name"].(string)
				modelMapper.RemoveModelInfo(unitType, dataGranularity,
					fmt.Sprintf("%s/%s", podNS, podName))
			}
		}
	}
}
