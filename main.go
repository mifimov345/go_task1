package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL    = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval = 5 * time.Second
)

func main() {
	errorCount := 0
	errorReported := false
	for {
		ok := fetchAndProcessOnce()
		if ok {
			errorCount = 0
			errorReported = false
		} else {
			errorCount++
		}
		if errorCount >= 3 && !errorReported {
			fmt.Println("Unable to fetch server statistic.")
			errorReported = true
		}
		time.Sleep(pollInterval)
	}
}

func fetchAndProcessOnce() bool {
	resp, err := http.Get(serverURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return false
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	line := strings.TrimSpace(string(bodyBytes))
	if line == "" {
		return false
	}
	parts := strings.Split(line, ",")
	if len(parts) != 7 {
		return false
	}
	values := make([]float64, 7)
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		num, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return false
		}
		values[i] = num
	}
	evaluateAndPrint(values)

	return true
}

func evaluateAndPrint(vals []float64) {
	loadAvg := vals[0]
	totalMemBytes := vals[1]
	usedMemBytes := vals[2]
	totalDiskBytes := vals[3]
	usedDiskBytes := vals[4]
	totalNetBytesPerSec := vals[5]
	usedNetBytesPerSec := vals[6]
	if loadAvg > 30 {
		fmt.Printf("Load Average is too high: %g\n", loadAvg)
	}
	if totalMemBytes > 0 {
		memUsage := usedMemBytes / totalMemBytes
		if memUsage > 0.8 {
			percent := int(memUsage * 100)
			fmt.Printf("Memory usage too high: %d%%\n", percent)
		}
	}
	if totalDiskBytes > 0 {
		diskUsage := usedDiskBytes / totalDiskBytes
		if diskUsage > 0.9 {
			freeBytes := totalDiskBytes - usedDiskBytes
			if freeBytes < 0 {
				freeBytes = 0
			}
			freeMb := int(freeBytes / (1024.0 * 1024.0))
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
		}
	}
	if totalNetBytesPerSec > 0 {
		netUsage := usedNetBytesPerSec / totalNetBytesPerSec
		if netUsage > 0.9 {
			freeBytesPerSec := totalNetBytesPerSec - usedNetBytesPerSec
			if freeBytesPerSec < 0 {
				freeBytesPerSec = 0
			}
			freeBitsPerSec := freeBytesPerSec * 8
			freeMbitPerSec := int(freeBitsPerSec / (1024.0 * 1024.0))
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbitPerSec)
		}
	}
}
