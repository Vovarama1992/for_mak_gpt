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
	"github.com/Vovarama1992/make_ziper/internal/classes"
	"github.com/Vovarama1992/make_ziper/internal/delivery"
	"github.com/Vovarama1992/make_ziper/internal/doc"
	"github.com/Vovarama1992/make_ziper/internal/domain"
	"github.com/Vovarama1992/make_ziper/internal/error_notificator"
	"github.com/Vovarama1992/make_ziper/internal/infra"
	"github.com/Vovarama1992/make_ziper/internal/minutes_packages"
	"github.com/Vovarama1992/make_ziper/internal/pdf"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/Vovarama1992/make_ziper/internal/speech"
	"github.com/Vovarama1992/make_ziper/internal/telegram"
	"github.com/Vovarama1992/make_ziper/internal/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {

	// =========================================================================
	// ENV / DB INIT
	// =========================================================================

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

	// =========================================================================
	// INFRASTRUCTURE
	// =========================================================================

	s3Client, err := infra.NewS3Client()
	if err != nil {
		log.Fatalf("failed to init s3: %v", err)
	}

	pdfConverter := pdf.NewPopplerPDFConverter()
	docConverter := doc.NewPythonDocConverter()

	// =========================================================================
	// REPOSITORIES
	// =========================================================================

	recordRepo := infra.NewRecordRepo(db)
	subscriptionRepo := infra.NewSubscriptionRepo(db)
	tariffRepo := infra.NewTariffRepo(db)
	botRepo := bots.NewRepo(db)
	minutePackageRepo := minutes_packages.NewMinutePackageRepo(db)
	classRepo := classes.NewClassRepo(db)
	userRepo := user.NewInfra(db)
	var authRepo ports.AuthRepo = infra.NewAuthRepo(db)

	// =========================================================================
	// ERROR NOTIFICATION
	// =========================================================================

	errInfra := error_notificator.NewInfra(nil)
	errService := error_notificator.NewService(errInfra)

	// =========================================================================
	// CLIENTS (AI / TTS / OCR etc.)
	// =========================================================================

	openAIClient := ai.NewOpenAIClient()
	ttsClient := speech.NewElevenLabsClient()

	// =========================================================================
	// DOMAIN SERVICES
	// =========================================================================

	s3Service := domain.NewS3Service(s3Client, errService)
	botService := bots.NewService(botRepo, s3Service)

	pdfService := pdf.NewPDFService(pdfConverter)
	docService := doc.NewService(docConverter)
	authService := domain.NewAuthService(authRepo, os.Getenv("AUTH_SECRET"))

	tariffService := domain.NewTariffService(tariffRepo)
	minutePackageService := minutes_packages.NewService(minutePackageRepo)
	classService := classes.NewClassService(classRepo)
	userService := user.NewService(userRepo)

	recordService := domain.NewRecordService(recordRepo, errService)

	speechService := speech.NewService(
		openAIClient, // Whisper
		ttsClient,    // ElevenLabs
		botService,
		errService,
	)

	aiService := ai.NewAiService(
		openAIClient,
		recordService,
		botRepo,
		classService,
		errService,
	)

	subscriptionService := domain.NewSubscriptionService(
		subscriptionRepo,
		tariffRepo,
		minutePackageService,
		errService,
	)

	// =========================================================================
	// TELEGRAM BOTS
	// =========================================================================

	botApp := telegram.NewBotApp(
		subscriptionService,
		tariffService,
		minutePackageService,
		aiService,
		speechService,
		recordService,
		s3Service,
		botService,
		userService,
		errService,
		classService,
		*pdfService,
		*docService,
	)

	if err := botApp.InitBots(); err != nil {
		log.Fatalf("failed to init telegram bots: %v", err)
	}

	errInfra.SetBots(botApp.GetBots())

	// =========================================================================
	// HTTP ROUTER
	// =========================================================================

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	// HANDLERS
	recordHandler := delivery.NewRecordHandler(recordService, zl)
	subHandler := delivery.NewSubscriptionHandler(subscriptionService)
	tariffHandler := delivery.NewTariffHandler(tariffService)
	botHandler := bots.NewHandler(botService)
	minPkgHandler := delivery.NewMinutePackageHandler(minutePackageService)
	classHandler := delivery.NewClassHandler(classService)
	authHandler := delivery.NewAuthHandler(authService)

	// ROUTES
	delivery.RegisterRoutes(
		r,
		recordHandler,
		subHandler,
		tariffHandler,
		botHandler,
		minPkgHandler,
		classHandler,
		authHandler,
	)

	r.With(httputil.RecoverMiddleware).Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	})

	// =========================================================================
	// BACKGROUND JOBS
	// =========================================================================

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx := context.Background()
			if err := subscriptionService.CleanupPending(ctx, 5*time.Minute); err != nil {
				log.Printf("[cleanup-pending] error: %v", err)
			} else {
				log.Printf("[cleanup-pending] removed old pending subscriptions")
			}
		}
	}()

	// =========================================================================
	// START SERVER
	// =========================================================================

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
