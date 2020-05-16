// The notexfr command provides a CLI for various data management operations for
// note-taking services.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rafaelespinoza/notexfr/internal/cmd"
)

func init() {
	cmd.Init()
}

func main() {
	if err := cmd.Run(context.Background()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
