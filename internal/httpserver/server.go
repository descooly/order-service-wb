package httpserver

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/descooly/order-service-wb/internal/service"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	srv *http.Server
}

func NewServer(svc *service.OrderService) *Server {
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal("Failed to create static sub-FS:", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			data, err := staticFiles.ReadFile("static/index.html")
			if err != nil {
				log.Printf("Failed to read index.html: %v", err)
				http.Error(w, "Page not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(data)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/order", svc.HTTPHandler)

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("HTTP server configured on %s", addr)
	return &Server{srv: server}
}

func (s *Server) Start() {
	go func() {
		log.Println("HTTP server starting...")
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	return s.srv.Shutdown(ctx)
}
