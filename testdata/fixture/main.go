// Command fixture is the Go port of test/_fixture.js.
//
// It is run as a subprocess with its output piped, so color detection should
// find no support and print both words unstyled. It lives under testdata so
// that it is not part of the module's package list.
package main

import (
	"fmt"

	"github.com/chalk/chalk-go"
)

func main() {
	fmt.Printf("%s %s\n",
		chalk.Hex("#ff6159").Sprint("testout"),
		chalk.Stderr.Hex("#ff6159").Sprint("testerr"))
}
