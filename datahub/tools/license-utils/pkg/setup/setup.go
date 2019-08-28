package setup

var (
	datahubAddress *string
)

func SetupInit(address *string) {
	datahubAddress = address
}

func SetDatahubAddress() error {
	// Input address
	address, err := InputAddress()
	if err != nil {
		return err
	}

	*datahubAddress = address

	return nil
}
