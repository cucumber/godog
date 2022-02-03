package main

import (
	"fmt"
	"os"

	"github.com/cucumber/godog/cmd/godog/internal"
)

func main() {
	rootCmd := internal.CreateRootCmd()
	buildCmd := internal.CreateBuildCmd()
	runCmd := internal.CreateRunCmd()
	versionCmd := internal.CreateVersionCmd()

	rootCmd.AddCommand(&buildCmd, &runCmd, &versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
