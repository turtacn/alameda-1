## Configuration

### The AI dispatcher configuration file
The default configuration file path is /etc/alameda/ai-dispatcher/ai-dispatcher.toml.

### AI dispatcher environment variables
AI dispatcher can use environment variables to override properties in the configuration file.

#### Mapping properties to environment variables
AI dispatcher specific environment variables are begin with token "ALAMEDA_AI_DISPATCHER_".Properties in the configuration file tree are seperated with underscore("_").And properties with dashes("-") are replaced with underscore.

##### Examples
* ALAMEDA_AI_DISPATCHER_DATAHUB_ADDRESS: The address of datahub.
* ALAMEDA_AI_DISPATCHER_QUEUE_URL: The url of queue.
* ALAMEDA_AI_DISPATCHER_MODEL_ENABLED: The flag of sending model jobs.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_24H_PREDICTIONJOBSENDINTERVALSEC: The interval of sending predicted jobs for granularity 24h.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_6H_PREDICTIONJOBSENDINTERVALSEC: The interval of sending predicted jobs for granularity 6h.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_1H_PREDICTIONJOBSENDINTERVALSEC: The interval of sending predicted jobs for granularity 1h.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_30S_PREDICTIONJOBSENDINTERVALSEC: The interval of sending predicted jobs for granularity 30s.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_24H_MODELJOBSENDINTERVALSEC: The interval of sending model jobs for granularity 24h.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_6H_MODELJOBSENDINTERVALSEC: The interval of sending model jobs for granularity 6h.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_1H_MODELJOBSENDINTERVALSEC: The interval of sending model jobs for granularity 1h.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_30S_MODELJOBSENDINTERVALSEC: The interval of sending model jobs for granularity 30s.
* ALAMEDA_AI_DISPATCHER_LOG_OUTPUTLEVEL: The output level of log.
* ALAMEDA_AI_DISPATCHER_MEASUREMENTS_CURRENT: The current measurement of prediction.
* ALAMEDA_AI_DISPATCHER_MEASUREMENTS_MAPE_THRESHOLD: The threshold of MAPE to send model jobs.
* ALAMEDA_AI_DISPATCHER_MEASUREMENTS_RMSE_THRESHOLD: The threshold of RMSE to send predicted jobs.
* ALAMEDA_AI_DISPATCHER_MEASUREMENTS_MINIMUMDATAPOINTS: The minimum number of data points to evaluate prediction.
* ALAMEDA_AI_DISPATCHER_HOURLYPREDICT: The flag of sending model and predicted jobs for granularity 30s except of VPA scaling tool.
* ALAMEDA_AI_DISPATCHER_SERVICESETTING_GRANULARITIES: Granularity types can be sent to queue
* ALAMEDA_AI_DISPATCHER_SERVICESETTING_PREDICTUNITS: Predict unit types can be sent to queue