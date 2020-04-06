// The snbackfill command provides a CLI for various metadata management
// operations in StandardNotes.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rafaelespinoza/snbackfill/internal/cmd"
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
