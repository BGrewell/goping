package main

import (
	"fmt"
	"github.com/BGrewell/goping/pinger"
	"sync"
	"time"
)

func main() {

	pingCount := 10
	pingInterval := 5 * time.Millisecond
	pingTimeout := 1 * time.Second

	var wg sync.WaitGroup
	outputFormat := "%-8d %-8v\n"
	for i := 0; i < pingCount; i++ {
		wg.Add(1)
		go func() {
			seq := i
			_, rtt, err := pinger.Ping("4.2.2.1", pingTimeout)
			if err != nil {
				fmt.Printf(outputFormat, seq, err)
			} else {
				fmt.Printf(outputFormat, seq, rtt)
			}
			wg.Done()
		}()
		// This scheduler releases payloads at fixed intervals.
		next := time.Now().Add(pingInterval)
		for time.Now().Before(next) {
			remaining := next.Sub(time.Now())
			if remaining > 100000 {
				time.Sleep(remaining / 4)
			}
			// spin for now.
		}
	}

	wg.Wait()
}
