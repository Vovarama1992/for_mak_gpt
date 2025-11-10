package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Vovarama1992/go-utils/httputil"
	"github.com/Vovarama1992/go-utils/logger"
	"github.com/Vovarama1992/make_ziper/internal/ai"
	"github.com/Vovarama1992/make_ziper/internal/delivery"
	"github.com/Vovarama1992/make_ziper/internal/domain"
	"github.com/Vovarama1992/make_ziper/internal/infra"
	"github.com/Vovarama1992/make_ziper/internal/telegram"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}
	defer db.Close()

	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync()
	zl := logger.NewZapLogger(baseLogger.Sugar())

	// --- Repos ---
	recordRepo := infra.NewRecordRepo(db)
	subscriptionRepo := infra.NewSubscriptionRepo(db)
	tariffRepo := infra.NewTariffRepo(db)

	// --- External clients ---
	s3Client, err := infra.NewS3Client()
	if err != nil {
		log.Fatalf("failed to init s3 client: %v", err)
	}

	// --- Services ---
	s3Service := domain.NewS3Service(s3Client)
	recordService := domain.NewRecordService(recordRepo, s3Service)
	subscriptionService := domain.NewSubscriptionService(subscriptionRepo, tariffRepo)
	tariffService := domain.NewTariffService(tariffRepo)

	// --- Handlers ---
	recordHandler := delivery.NewRecordHandler(recordService, zl)
	subscriptionHandler := delivery.NewSubscriptionHandler(subscriptionService)
	tariffHandler := delivery.NewTariffHandler(tariffService)

	// --- AI module ---
	aiClient := ai.NewOpenAIClient()
	aiService := ai.NewAiService(aiClient, recordService)

	// --- Telegram bots initialization ---
	botApp := &telegram.BotApp{
		SubscriptionService: subscriptionService,
		TariffService:       tariffService,
		AiService:           aiService,
	}
	if err := botApp.InitBots(); err != nil {
		log.Fatalf("failed to init telegram bots: %v", err)
	}

	// --- Router ---
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	delivery.RegisterRoutes(r, recordHandler, subscriptionHandler, tariffHandler)

	r.With(httputil.RecoverMiddleware).Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	addr := ":" + port
	zl.Log(logger.LogEntry{
		Level:   "info",
		Message: "listening at " + addr,
		Service: "make_ziper",
	})

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
