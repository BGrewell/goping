package main

import (
	"flag"
	"fmt"
	"github.com/BGrewell/goping/pinger"
	"github.com/fatih/color"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	version string = "0.1.0"
	build   string
)

func PrintUsageLine(parameter string, defaultValue interface{}, description string, extra string) {
	yellow := color.New(color.FgHiYellow)
	cyan := color.New(color.FgHiCyan)
	red := color.New(color.FgHiRed)
	yellow.Printf("    %-22s", parameter)
	cyan.Printf("  %-14v", defaultValue)
	yellow.Printf("  %-36s", description)
	red.Printf("  %s\n", extra)
}

func Usage() (usage func()) {
	return func() {
		white := color.New(color.FgWhite)
		boldWhite := color.New(color.FgWhite, color.Bold)
		boldGreen := color.New(color.FgGreen, color.Bold)
		usageLineFormat := "    %-22s  %-14v  %s\n"
		boldGreen.Printf("[+] goping :: Version %v :: Build %v\n", version, build)
		boldWhite.Print("Usage: ")
		fmt.Printf("goping <flags> target\n")
		boldGreen.Print("  General Options:\n")
		white.Printf(usageLineFormat, "Parameter", "Default", "Description")
		//yellow.Printf(usageLineFormat, "-h", false, "show this help output")
		PrintUsageLine("--h[elp]", false, "show this help output", "")
		PrintUsageLine("--json", false, "output machine readable json", "[not implemented]")
		boldGreen.Printf("  General Rule Parameters:\n")
		white.Printf(usageLineFormat, "Parameter", "Default", "Description")
		PrintUsageLine("--count", 4, "number of pings to send", "")
		PrintUsageLine("--async", false, "send pings asynchronously", "[flag]")
		PrintUsageLine("--interval", 1.0, "milliseconds between pings", "")
		PrintUsageLine("--timeout", 1000, "milliseconds to wait ping replies", "")
	}
}

func main() {

	var count = flag.Int("count", 4, "")
	var async = flag.Bool("async", false, "")
	var interval = flag.Float64("interval", 1.0, "")
	var timeoutMs = flag.Int("timeout", 1000, "")
	flag.Usage = Usage()
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("target is a required positional argument")
		flag.Usage()
		os.Exit(1)
	}
	var target = flag.Args()[0]
	timeout := time.Duration(*timeoutMs) * time.Millisecond
	var wg sync.WaitGroup

	for i := 0; i < *count; i++ {
		wg.Add(1)
		go func(seq int) {
			_, sendtime, rtt, err := pinger.Ping(target, uint16(seq), timeout)
			if err != nil {
				fmt.Printf("%-8d  %20d  %-8v\n", seq, sendtime, err)
			} else {
				output := strings.Replace(fmt.Sprintf("%-8d  %20d  %8.12v\n", seq, sendtime, rtt), "ms", "", 1)
				fmt.Printf(output)
			}
			wg.Done()
		}((i+1) % int(^uint16(0)))

		// This scheduler releases payloads at fixed intervals.
		next := time.Now().Add(time.Duration(*interval) * time.Millisecond)
		for time.Now().Before(next) {
			remaining := next.Sub(time.Now())
			if remaining > 100000 {
				time.Sleep(remaining / 4)
			}
			// spin for now.
		}

		// If we are synchronous then wait for that routine to finish otherwise we just start another
		if !*async {
			wg.Wait()
		}

	}

	wg.Wait()
}
