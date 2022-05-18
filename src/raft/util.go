package raft

import (
	"log"
	"math/rand"
	"time"
)

// Debugging
const Debug = true

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

// RandomTime return a random time between min and max
func RandomTime() time.Duration {
	return time.Duration(rand.Intn(6))*time.Second/10 + time.Second*5/10
}
