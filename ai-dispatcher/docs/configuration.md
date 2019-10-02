## Configuration

### The AI dispatcher configuration file
The default configuration file path is /etc/alameda/ai-dispatcher/ai-dispatcher.toml.

### AI dispatcher environment variables
AI dispatcher can use environment variables to override properties in the configuration file.

#### Mapping properties to environment variables
AI dispatcher specific environment variables are begin with token "ALAMEDA_AI_DISPATCHER_".Properties in the configuration file tree are seperated with underscore("_").And properties with dashes("-") are replaced with underscore.

##### Examples
* ALAMEDA_AI_DISPATCHER_DATAHUB_ADDRESS: Can be used to override property datahub addresss in configuration file.
* ALAMEDA_AI_DISPATCHER_QUEUE_URL: Can be used to override property queue url in configuration file.
* ALAMEDA_AI_DISPATCHER_MODEL_ENABLED: Can be used to override property model job sent enabled in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_24H_PREDICTIONJOBSENDINTERVALSEC: Can be used to override property predict job sent interval with 24h granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_6H_PREDICTIONJOBSENDINTERVALSEC: Can be used to override property predict job sent interval with 6h granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_1H_PREDICTIONJOBSENDINTERVALSEC: Can be used to override property predict job sent interval with 1h granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_30S_PREDICTIONJOBSENDINTERVALSEC: Can be used to override property predict job sent interval with 30s granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_24H_MODELJOBSENDINTERVALSEC: Can be used to override property model job sent interval with 24h granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_6H_MODELJOBSENDINTERVALSEC: Can be used to override property model job sent interval with 6h granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_1H_MODELJOBSENDINTERVALSEC: Can be used to override property model job sent interval with 1h granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_GRANULARITIES_30S_MODELJOBSENDINTERVALSEC: Can be used to override property model job sent interval with 30s granularity in configuration file.
* ALAMEDA_AI_DISPATCHER_LOG_OUTPUTLEVEL: Can be used to override property log output level in configuration file.
* ALAMEDA_AI_DISPATCHER_MEASUREMENTS_CURRENT: Can be used to override property error measurement used in configuration file.
* ALAMEDA_AI_DISPATCHER_MEASUREMENTS_MAPE_THRESHOLD: Can be used to override property MAPE threshold used in configuration file.
* ALAMEDA_AI_DISPATCHER_MEASUREMENTS_RMSE_THRESHOLD: Can be used to override property RMSE threshold used in configuration file.
