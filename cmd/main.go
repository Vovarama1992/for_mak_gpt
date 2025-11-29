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
	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/delivery"
	"github.com/Vovarama1992/make_ziper/internal/domain"
	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/infra"
	"github.com/Vovarama1992/make_ziper/internal/minutes_packages"
	"github.com/Vovarama1992/make_ziper/internal/speech"
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

	// === repos ===
	recordRepo := infra.NewRecordRepo(db)
	subscriptionRepo := infra.NewSubscriptionRepo(db)
	tariffRepo := infra.NewTariffRepo(db)
	botRepo := bots.NewRepo(db)
	minutePackageRepo := minutes_packages.NewMinutePackageRepo(db)

	// === S3 ===
	s3Client, err := infra.NewS3Client()
	if err != nil {
		log.Fatalf("failed to init s3 client: %v", err)
	}

	// === clients ===
	from_speech := ai.NewOpenAIClient()
	to_speech := speech.NewElevenLabsClient()
	aiClient := ai.NewOpenAIClient()
	botService := bots.NewService(botRepo)

	// === services ===
	tariffService := domain.NewTariffService(tariffRepo)
	minutePackageService := minutes_packages.NewService(minutePackageRepo)

	// === error notificator ===
	errorInfra := error_notificator.NewInfra(nil)
	errorService := error_notificator.NewService(errorInfra)

	// === speech ===
	speechService := speech.NewService(
		from_speech,
		to_speech,
		botService,
		errorService,
	)

	// === S3 + record ===
	s3Service := domain.NewS3Service(s3Client, errorService)
	recordService := domain.NewRecordService(recordRepo, errorService)

	// === AI ===
	aiService := ai.NewAiService(aiClient, recordService, botRepo, errorService)

	// === subscriptions ===
	subscriptionService := domain.NewSubscriptionService(
		subscriptionRepo,
		tariffRepo,
		minutePackageService,
		errorService,
	)

	// === telegram bots ===
	botApp := telegram.NewBotApp(
		subscriptionService,
		tariffService,
		minutePackageService,
		aiService,
		speechService,
		recordService,
		s3Service,
		botService,
		errorService,
	)

	if err := botApp.InitBots(); err != nil {
		log.Fatalf("failed to init telegram bots: %v", err)
	}

	errorInfra.SetBots(botApp.GetBots())

	// === HTTP server ===
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	recordHandler := delivery.NewRecordHandler(recordService, zl)
	subscriptionHandler := delivery.NewSubscriptionHandler(subscriptionService)
	tariffHandler := delivery.NewTariffHandler(tariffService)
	promptHandler := bots.NewHandler(botService)
	minutePackageHandler := delivery.NewMinutePackageHandler(minutePackageService)

	delivery.RegisterRoutes(
		r,
		recordHandler,
		subscriptionHandler,
		tariffHandler,
		promptHandler,
		minutePackageHandler,
	)

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
