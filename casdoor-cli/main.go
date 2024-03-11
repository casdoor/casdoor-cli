/*
Copyright © 2024 Fabien CHEVALIER
*/
package main

import (
	"gitlab.com/sdv9972401/casdoor-cli-go/cmd"
	_ "gitlab.com/sdv9972401/casdoor-cli-go/cmd/users"
)

var appVersion string //set on build

func main() {
	cmd.Execute()
}
