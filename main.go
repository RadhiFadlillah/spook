package main

import (
	"github.com/RadhiFadlillah/spook/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	spookCmd := cmd.NewSpookCmd()
	if err := spookCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
