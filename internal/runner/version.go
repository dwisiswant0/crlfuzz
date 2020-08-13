package runner

import (
	"fmt"
	"os"
)

func showVersion() {
	fmt.Printf("CRLFuzz %s\n", version)
	os.Exit(2)
}
