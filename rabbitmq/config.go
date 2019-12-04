package rabbitmq

type RabbitmqConfig struct {
	Account  string
	Password string
	Address  string
	Port     string
}

func NewRabbitmqConfig(account string, password string, address string, port string) *RabbitmqConfig {
	return &RabbitmqConfig{
		Account:  account,
		Password: password,
		Address:  address,
		Port:     port,
	}
}
