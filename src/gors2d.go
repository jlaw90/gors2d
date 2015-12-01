package main

import (
	"fmt"
	"time"
	"rs2d/server/jaggrab"
	"rs2d/server"
	"rs2d/config"
	"rs2d/cache"
)

func main() {
	fmt.Println("RS2D-GO starting up...")

	// Load configuration files
	config.Load()

	// Initialise the cache (needed for JAGGRAB and update server)
	cache.Initialise(config.Config.CachePath)

	// Start the JAGGRAB server
	jaggrab.Start()
	// Start the main server
	server.Start()

	// Infinite loop, could do maintenance tasks here or something...
	for {
		time.Sleep(1000)
	}
}
