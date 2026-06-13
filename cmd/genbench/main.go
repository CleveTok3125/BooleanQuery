package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

var levels = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
var components = []string{"server", "database", "router", "cache", "worker", "api", "auth", "scheduler"}
var messages = []string{
	"connection timeout after %ds (host=%s, db=%s)",
	"request processed in %dms (status=%d, path=%s)",
	"memory usage at %d%% (threshold=%d%%)",
	"failed to connect to %s:%d (attempt %d/%d)",
	"user login from IP %s (user_id=%d, method=%s)",
	"disk I/O latency: %dms (device=%s, queue=%d)",
	"cache miss for key=%s (table=%s, ttl=%d)",
	"replication lag: %dms (source=%s, target=%s)",
	"task queue depth: %d (worker=%s, max=%d)",
}

func randIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func randPath() string {
	paths := []string{"/api/v1/users", "/api/v1/orders", "/health", "/metrics", "/static/js/app.js"}
	return paths[rand.Intn(len(paths))]
}

func main() {
	const numLines = 100000
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	f, err := os.Create("testdata/bench.log")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for range numLines {
		level := levels[rng.Intn(len(levels))]
		comp := components[rng.Intn(len(components))]
		msg := messages[rng.Intn(len(messages))]
		var line string

		switch msg {
		case messages[0]:
			line = fmt.Sprintf(msg, rng.Intn(60), randIP(), comp)
		case messages[1]:
			line = fmt.Sprintf(msg, rng.Intn(500), []int{200, 201, 301, 400, 403, 404, 500}[rng.Intn(7)], randPath())
		case messages[2]:
			line = fmt.Sprintf(msg, rng.Intn(100), 90)
		case messages[3]:
			line = fmt.Sprintf(msg, comp, rng.Intn(10000), rng.Intn(5)+1, 5)
		case messages[4]:
			line = fmt.Sprintf(msg, randIP(), rng.Intn(100000), []string{"oauth", "password", "token"}[rng.Intn(3)])
		case messages[5]:
			line = fmt.Sprintf(msg, rng.Intn(2000), []string{"sda", "sdb", "nvme0"}[rng.Intn(3)], rng.Intn(256))
		case messages[6]:
			line = fmt.Sprintf(msg, fmt.Sprintf("key_%d", rng.Intn(1000)), comp, rng.Intn(3600))
		case messages[7]:
			line = fmt.Sprintf(msg, rng.Intn(5000), comp, comp)
		case messages[8]:
			line = fmt.Sprintf(msg, rng.Intn(1000), comp, 100)
		default:
			line = "unknown message"
		}

		ts := time.Date(2026, 6, 13, rng.Intn(24), rng.Intn(60), rng.Intn(60), 0, time.UTC).Format(time.RFC3339)
		fmt.Fprintf(f, "%s [%s] [%s] %s\n", ts, level, comp, line)
	}

	log.Printf("Generated %d lines to testdata/bench.log", numLines)
}
