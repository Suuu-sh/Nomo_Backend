package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/yota/nomo/backend/internal/config"
	"github.com/yota/nomo/backend/internal/httpapi"
	"github.com/yota/nomo/backend/internal/supabase"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration error", "error", err)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: 15 * time.Second}
	supabaseClient := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseAnonKey, httpClient)
	var adminSupabaseClient *supabase.Client
	if cfg.SupabaseServiceRoleKey != "" {
		adminSupabaseClient = supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey, httpClient)
	}
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(httpapi.Dependencies{Config: cfg, Logger: logger, Supabase: supabaseClient, AdminSupabase: adminSupabaseClient}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting nomo backend", "port", cfg.Port, "env", cfg.Environment)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
