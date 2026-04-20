package pricectl

import (
	"embed"
	"fmt"
	"net/http"
	"os"
)

//go:embed ui/index.html
var uiFiles embed.FS

func CmdServe() {
	port := "8080"

	mux := http.NewServeMux()

	// Serve the UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := uiFiles.ReadFile("ui/index.html")
		if err != nil {
			http.Error(w, "ui not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	// API endpoints
	mux.HandleFunc("/api/products", apiProducts)
	mux.HandleFunc("/api/check", apiCheck)
	mux.HandleFunc("/api/history", apiHistory)

	addr := "127.0.0.1:" + port
	fmt.Printf("pricectl UI running at http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
