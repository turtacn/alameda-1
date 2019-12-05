package main

import (
	"os"

	rabbit_app "github.com/containers-ai/alameda/rabbitmq/cmd/app"
)

func main() {
	if err := rabbit_app.PublishCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
