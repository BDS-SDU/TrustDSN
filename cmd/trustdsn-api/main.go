package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("get working directory: %v", err)
	}

	if envRoot := os.Getenv("TRUSTDSN_REPO_ROOT"); envRoot != "" {
		repoRoot = envRoot
	}

	addr := os.Getenv("TRUSTDSN_API_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8080"
	}

	minerHTTPPort := 18080
	if envPort := os.Getenv("TRUSTDSN_MINER_HTTP_PORT"); envPort != "" {
		p, err := strconv.Atoi(envPort)
		if err != nil {
			log.Fatalf("invalid TRUSTDSN_MINER_HTTP_PORT: %v", err)
		}
		minerHTTPPort = p
	}

	cfg := Config{
		Addr:          addr,
		RepoRoot:      repoRoot,
		LotusBin:      "./lotus",
		MinerHTTPPort: minerHTTPPort,
	}

	srv := NewServer(cfg)

	log.Printf("trustdsn-api listening on %s", cfg.Addr)
	log.Printf("repo root: %s", cfg.RepoRoot)
	log.Printf("miner http port: %d", cfg.MinerHTTPPort)

	if err := http.ListenAndServe(cfg.Addr, srv.Routes()); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
