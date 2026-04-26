package main

import (
	"log"

	"github.com/NirajDonga/dbpods/tcp-proxy/internal/config"
	"github.com/NirajDonga/dbpods/tcp-proxy/internal/proxy"
)

func main() {
	cfg := config.Load()

	p := proxy.NewProxy(":"+cfg.ProxyPort, cfg.Namespace)

	log.Printf("[Proxy] Starting on :%s (namespace: %s)", cfg.ProxyPort, cfg.Namespace)
	if err := p.Start(); err != nil {
		log.Fatalf("[Proxy] Fatal: %v", err)
	}
}
