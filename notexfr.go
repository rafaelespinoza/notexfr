// The notexfr command provides a CLI for various data management operations for
// note-taking services.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rafaelespinoza/notexfr/internal/cmd"
)

func main() {
	err := cmd.New().ExecuteContext(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
