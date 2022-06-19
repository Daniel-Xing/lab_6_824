package raft

import (
	"log"
	"math/rand"
	"time"
)

// Debugging
const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

// RandomTime return a random time between min and max
func RandomTime() time.Duration {
	return time.Duration(rand.Intn(1000))*time.Millisecond + time.Millisecond*1000
}

//
func min(num1, num2 int) int {
	if num1 > num2 {
		return num2
	}
	return num1
}
