package main

import (
	"os"

	"github.com/sftpgo/sftpgo-plugin-eventstore/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
