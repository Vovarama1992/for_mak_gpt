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
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Vovarama1992/make_ziper/internal/ai"
	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/classes"
	"github.com/Vovarama1992/make_ziper/internal/delivery"
	"github.com/Vovarama1992/make_ziper/internal/doc"
	"github.com/Vovarama1992/make_ziper/internal/domain"
	"github.com/Vovarama1992/make_ziper/internal/infra"
	"github.com/Vovarama1992/make_ziper/internal/minutes_packages"
	error_notificator "github.com/Vovarama1992/make_ziper/internal/notificator"
	"github.com/Vovarama1992/make_ziper/internal/pdf"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/Vovarama1992/make_ziper/internal/speech"
	"github.com/Vovarama1992/make_ziper/internal/telegram"
	"github.com/Vovarama1992/make_ziper/internal/textrules"
	"github.com/Vovarama1992/make_ziper/internal/trial"
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

	// ⬇️ ТОЛЬКО ДЛЯ БД
	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(dbCtx); err != nil {
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
	trialRepo := trial.NewRepo(db)
	var authRepo ports.AuthRepo = infra.NewAuthRepo(db)

	textRuleRepo := textrules.NewRepo(db)

	// =========================================================================
	// ERROR NOTIFICATION
	// =========================================================================

	errInfra := error_notificator.NewInfra(nil)
	errService := error_notificator.NewService(errInfra)

	// =========================================================================
	// CLIENTS
	// =========================================================================

	openAIClient := ai.NewOpenAIClient()
	perplexityTTS := speech.NewPerplexityTTS()
	ttsClient := speech.NewElevenLabsClient()
	perplexityClient := ai.NewPerplexityClient()
	paymentProvider := infra.NewYooKassaProvider()

	// =========================================================================
	// DOMAIN SERVICES
	// =========================================================================

	s3Service := domain.NewS3Service(s3Client, errService)
	botService := bots.NewService(botRepo, s3Service)

	pdfService := pdf.NewPDFService(pdfConverter)
	docService := doc.NewService(docConverter)
	authService := domain.NewAuthService(authRepo, os.Getenv("AUTH_SECRET"))

	tariffService := domain.NewTariffService(tariffRepo)
	minutePackageService := minutes_packages.NewService(
		minutePackageRepo,
		paymentProvider,
	)
	classService := classes.NewClassService(classRepo)
	userService := user.NewService(userRepo)

	recordService := domain.NewRecordService(recordRepo, errService)

	speechService := speech.NewService(
		openAIClient,
		ttsClient,
		botService,
		errService,
	)

	aiService := ai.NewAiService(
		openAIClient,
		perplexityClient,
		recordService,
		botRepo,
		classService,
		errService,
	)

	subscriptionService := domain.NewSubscriptionService(
		subscriptionRepo,
		tariffRepo,
		trialRepo,
		minutePackageService,
		errService,
		paymentProvider,
	)

	textRuleService := textrules.NewService(textRuleRepo)

	// =========================================================================
	// TELEGRAM BOTS
	// =========================================================================

	botApp := telegram.NewBotApp(
		subscriptionService,  // ports.SubscriptionService
		tariffService,        // ports.TariffService
		minutePackageService, // minutes_packages.MinutePackageService
		trialRepo,            // trial.RepoInf

		aiService,     // *ai.AiService
		speechService, // *speech.Service
		perplexityTTS, // *speech.PerplexityTTS

		textRuleService, // textrules.Service
		recordService,   // ports.RecordService
		s3Service,       // ports.S3Service
		botService,      // bots.Service
		userService,     // user.Service
		errService,      // notificator.Notificator
		classService,    // classes.ClassService
		*pdfService,     // pdf.PDFService
		*docService,     // doc.Service
	)

	botApp.SetAdminBotUsername(os.Getenv("ADMIN_BOT_USERNAME"))

	// ⬇️ ВАЖНО: БЕЗ TIMEOUT
	botCtx := context.Background()

	if err := botApp.InitBots(botCtx); err != nil {
		log.Fatalf("failed to init telegram bots: %v", err)
	}

	adminToken := os.Getenv("ADMIN_BOT_TOKEN")
	if err := botApp.InitAdminBot(botCtx, adminToken); err != nil {
		log.Fatalf("failed to init admin bot: %v", err)
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

	recordHandler := delivery.NewRecordHandler(recordService, zl)
	subHandler := delivery.NewSubscriptionHandler(subscriptionService)
	tariffHandler := delivery.NewTariffHandler(tariffService)
	botHandler := bots.NewHandler(botService)
	minPkgHandler := delivery.NewMinutePackageHandler(minutePackageService)
	classHandler := delivery.NewClassHandler(classService)
	authHandler := delivery.NewAuthHandler(authService)
	textRuleHandler := delivery.NewTextRuleHandler(textRuleRepo)

	delivery.RegisterRoutes(
		r,
		recordHandler,
		subHandler,
		tariffHandler,
		botHandler,
		minPkgHandler,
		classHandler,
		authHandler,
		textRuleHandler,
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

			// 1) чистим pending
			if err := subscriptionService.CleanupPending(ctx, 5*time.Minute); err != nil {
				log.Printf("[cleanup-pending] error: %v", err)
			}

			// 2) переводим все истёкшие active → expired
			expired, err := subscriptionRepo.ExpireDue(ctx)
			if err != nil {
				log.Printf("[expire] error: %v", err)
				continue
			}

			// 3) уведомляем ТОЛЬКО триалы соответствующего бота
			for _, sub := range expired {
				if sub.PlanID == nil {
					continue
				}

				// получаем trial для КОНКРЕТНОГО бота
				trial, err := tariffRepo.GetTrial(ctx, sub.BotID)
				if err != nil || trial == nil {
					continue
				}

				// если это не trial — пропускаем
				if *sub.PlanID != int64(trial.ID) {
					continue
				}

				// если уже уведомляли — пропускаем
				if sub.TrialNotifiedAt != nil {
					continue
				}

				bot, ok := botApp.GetBots()[sub.BotID]
				if !ok || bot == nil {
					continue
				}

				text := botApp.BuildSubscriptionText(ctx, sub.BotID)
				kb := botApp.BuildSubscriptionMenu(ctx, sub.BotID)

				msg := tgbotapi.NewMessage(sub.TelegramID, text)
				msg.ReplyMarkup = kb

				if _, err := bot.Send(msg); err != nil {
					continue
				}

				_ = subscriptionRepo.MarkTrialNotified(ctx, sub.ID)
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
