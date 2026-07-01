package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"platformgateway/internal/config"
	"platformgateway/internal/gateway"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "config file path")
	flag.Parse()
	abs, err := filepath.Abs(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Load(abs)
	if err != nil {
		log.Fatal(err)
	}
	engine := gateway.Setup(cfg)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("PlatformGateway listening on http://localhost%s", addr)
	log.Printf("  IAM -> %s", cfg.Upstreams.UserCore)
	log.Printf("  PIM -> %s (JWT validate=%v)", cfg.Upstreams.ProductCore, cfg.JWT.ValidatePIM)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
