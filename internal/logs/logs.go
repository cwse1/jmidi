// Package logs simply wraps the go/log package.
package logs

import (
	"fmt"
	"log"
)

// Fatal is a wrapper for log.Fatalf()
func Fatal(err error) {
	log.Fatalf("[FATAL]: %v", err)
}

// Warn is a wrapper for log.Printf()
func Warn(err error) {
	log.Printf("[WARN]: %v", err)
}

// CheckErr logs file closing errors and exits
func CheckErr(err error) {
	Fatal(fmt.Errorf("failed closing file: %v", err))
}
