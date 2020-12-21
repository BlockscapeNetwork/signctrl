package utils

import (
	"log"
	"net"
	"time"
)

// RetryDial keeps trying to dial in intervals.
func RetryDial(protocol, address string, logger *log.Logger) net.Conn {
	for {
		conn, err := net.Dial(protocol, address)
		if err == nil {
			logger.Printf("[INFO] Established connection with %v\n", address)
			return conn
		}

		logger.Printf("[ERR] pairmint: %v\n", err.Error())
		time.Sleep(time.Second)
	}
}
