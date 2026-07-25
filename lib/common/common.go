// Package common is the shared library of the example monorepo: one
// function, consumed by services/api through a replace directive.
package common

// Greeting returns the shared greeting the api service prints.
func Greeting() string {
	return "hello from lib/common"
}
