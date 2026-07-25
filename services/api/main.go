// Command api prints the shared greeting — the smallest possible consumer
// of a replace-directive sibling in a Go monorepo.
package main

import (
	"fmt"

	common "example.com/lib/common"
)

func main() {
	fmt.Println(common.Greeting())
}
