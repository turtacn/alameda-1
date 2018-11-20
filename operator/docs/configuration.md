## Configuration

### The Operator configuration file
The default configuration file path is /etc/alameda/operator/operator.yml.

### Operator environment variables
Alameda Operator can use environment variables to override properties in the configuration file.

#### Mapping properties to environment variables
Alameda Operator specific environment variables are begin with token "ALAMEDA_OPERATOR_".Properties in the configuration file tree are seperated with underscore("_").And properties with dashes("-") are replaced with underscore.

##### Examples
* ALAMEDA_OPERATOR_GRPC_BIND_ADDRESS: Can be used to override property bind-address under gRPC node in configuration file.