package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/WAT36/shorty/internal/httpserver"
	"github.com/WAT36/shorty/internal/store"
)

func main() {
	dbPath := getenv("SHORTY_DB", "data/urls.json")
	addr := ":" + getenv("PORT", "8080")

	// ストア初期化（JSONファイルへ永続化）
	s, err := store.NewFileStore(dbPath)
	if err != nil {
		log.Fatalf("store init error: %v", err)
	}
	if err := s.Load(); err != nil {
		log.Printf("no initial db found or load error: %v (will create on save)", err)
	}

	// HTTPサーバ
	srv := httpserver.New(addr, s)

	// サーバ起動
	go func() {
		log.Printf("shorty is running on http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	// 終了前に保存
	if err := s.Save(); err != nil {
		log.Printf("save error on shutdown: %v", err)
	}
	log.Println("bye!")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
