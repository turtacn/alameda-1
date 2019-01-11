## Configuration

### The Operator configuration file
The default configuration file path is /etc/alameda/operator/operator.yml.

### Operator environment variables
Federator.ai Operator can use environment variables to override properties in the configuration file.

#### Mapping properties to environment variables
Federator.ai Operator specific environment variables are begin with token "ALAMEDA_".Properties in the configuration file tree are seperated with underscore("_").And properties with dashes("-") are replaced with underscore.

##### Examples
* ALAMEDA_GRPC_BIND_ADDRESS: Can be used to override property bind-address under gRPC node in configuration file.
