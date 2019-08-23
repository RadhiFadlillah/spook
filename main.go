package main

import (
	"github.com/go-spook/spook/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.SpookCmd().Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
