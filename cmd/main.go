package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Khramov-Vladislav/digital-analytics/internal/handlers"
)

func main() {
	// Ensure fonts are available
	if err := ensureFonts(); err != nil {
		log.Printf("Warning: could not download fonts: %v", err)
		log.Printf("Using Helvetica fallback font")
	}

	addr := flag.String("addr", ":8080", "HTTP network address")
	flag.Parse()

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)

	log.Printf("Starting server on %s", *addr)
	err := http.ListenAndServe(*addr, mux)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
