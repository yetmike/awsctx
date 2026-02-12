package main

import (
	"fmt"
	"os"

	"github.com/yetmike/awsctx/internal/awsctx"
)

func main() {
	if err := awsctx.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
