package validator

import "fmt"

// Validate act as a middleware to validate the incoming request against the policy and the consent requirements.
func Validate() {
	fmt.Println("This is the main validator")
}
