/*
Copyright © 2025 Dmitriy Pyankov <mail@zerodi.ru>
*/
package main

import (
	"fmt"
	"os"

	"github.com/zerodi/cctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
