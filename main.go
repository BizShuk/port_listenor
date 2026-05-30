package main

import (
	"github.com/bizshuk/port_listenor/cmd"
	"github.com/bizshuk/port_listenor/config"
)

func main() {
	config.Default()
	cmd.Execute()
}
