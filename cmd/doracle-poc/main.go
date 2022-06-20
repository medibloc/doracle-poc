package main

import (
	"github.com/medibloc/doracle-poc/cmd/doracle-poc/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Error(err)
	}
}
