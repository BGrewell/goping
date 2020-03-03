package main

import (
	"fmt"
	"github.com/BGrewell/goping/pinger"
)

func main() {

	dest, rtt, err := pinger.Ping("4.2.2.1", 1000)
	if err != nil {
		fmt.Printf("error: %v", err)
	} else {
		fmt.Printf("reply from %v: %v\n", dest, rtt)
	}

}
