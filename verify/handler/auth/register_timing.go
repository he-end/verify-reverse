package auth

import (
	"crypto/rand"
	"math/big"
	"time"
)

const (
	registerMinDelay = 4000
	registerMaxDelay = 8000
)

func sleepRemaining(elapsed time.Duration) time.Duration {
	n, err := rand.Int(rand.Reader, big.NewInt(registerMaxDelay-registerMinDelay+1))
	if err != nil {
		n = big.NewInt(0)
	}
	delayMS := registerMinDelay + int(n.Int64())
	delay := time.Duration(delayMS) * time.Millisecond
	remaining := delay - elapsed
	if remaining > 0 {
		time.Sleep(remaining)
	}
	return delay
}
