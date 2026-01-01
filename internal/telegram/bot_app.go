package telegram

import (
	"context"
	"log"
	"time"

	"github.com/Vovarama1992/make_ziper/internal/ai"
	"github.com/Vovarama1992/make_ziper/internal/bots"
	"github.com/Vovarama1992/make_ziper/internal/classes"
	"github.com/Vovarama1992/make_ziper/internal/doc"
	mpkg "github.com/Vovarama1992/make_ziper/internal/minutes_packages"
	notificator "github.com/Vovarama1992/make_ziper/internal/notificator"
	"github.com/Vovarama1992/make_ziper/internal/pdf"
	"github.com/Vovarama1992/make_ziper/internal/ports"
	"github.com/Vovarama1992/make_ziper/internal/speech"
	"github.com/Vovarama1992/make_ziper/internal/textrules"
	"github.com/Vovarama1992/make_ziper/internal/user"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ==================================================
// ADMIN HELP CONTEXT
// ==================================================

type AdminHelpContext struct {
	BotID  string
	UserID int64
}

// ==================================================
// BOT APP
// ==================================================

type BotApp struct {
	SubscriptionService  ports.SubscriptionService
	TariffService        ports.TariffService
	MinutePackageService mpkg.MinutePackageService
	AiService            *ai.AiService
	SpeechService        *speech.Service
	TextRuleService      textrules.Service
	RecordService        ports.RecordService
	S3Service            ports.S3Service
	PDFService           pdf.PDFService
	DocService           doc.Service

	BotsService bots.Service
	UserService user.Service

	ErrorNotify notificator.Notificator

	bots          map[string]*tgbotapi.BotAPI
	shownKeyboard map[string]map[int64]bool

	ClassService classes.ClassService

	// user -> admin
	helpMode map[string]map[int64]bool

	// admin -> user (старое, оставлено)
	adminHelpMode map[int64]*AdminHelpContext

	// ✅ НОВОЕ: отдельный admin-бот
	adminBot         *AdminBot
	adminBotUsername string
}

// ==================================================
// CONSTRUCTOR
// ==================================================

func NewBotApp(
	subs ports.SubscriptionService,
	tariffs ports.TariffService,
	minutePkg mpkg.MinutePackageService,
	ai *ai.AiService,
	speech *speech.Service,
	textRules textrules.Service,
	record ports.RecordService,
	s3 ports.S3Service,
	bots bots.Service,
	userSvc user.Service,
	errNotify notificator.Notificator,
	classes classes.ClassService,
	pdf pdf.PDFService,
	doc doc.Service,
) *BotApp {
	if minutePkg == nil {
		panic("MinutePackageService is nil")
	}

	return &BotApp{
		SubscriptionService:  subs,
		TariffService:        tariffs,
		MinutePackageService: minutePkg,
		AiService:            ai,
		SpeechService:        speech,
		TextRuleService:      textRules,
		RecordService:        record,
		S3Service:            s3,
		PDFService:           pdf,
		DocService:           doc,

		BotsService: bots,
		UserService: userSvc,

		ErrorNotify:  errNotify,
		ClassService: classes,

		helpMode:      make(map[string]map[int64]bool),
		adminHelpMode: make(map[int64]*AdminHelpContext),
	}
}

// ==================================================
// INIT BOTS
// ==================================================

func (app *BotApp) InitBots(ctx context.Context) error {
	app.bots = make(map[string]*tgbotapi.BotAPI)
	app.shownKeyboard = make(map[string]map[int64]bool)

	cfgs, err := app.BotsService.ListAll(ctx)
	if err != nil {
		return err
	}

	for _, cfg := range cfgs {
		if cfg.Token == "" {
			continue
		}

		bot, err := tgbotapi.NewBotAPI(cfg.Token)
		if err != nil {
			log.Printf("[bot_app] init fail for %s: %v", cfg.BotID, err)
			continue
		}

		app.bots[cfg.BotID] = bot
		log.Printf("[bot_app] ready: @%s (%s)", bot.Self.UserName, cfg.BotID)

		go app.runBotLoop(cfg.BotID, bot)
	}

	app.startTrialCleanupTicker(ctx, 1*time.Minute)

	return nil
}

func (app *BotApp) GetBots() map[string]*tgbotapi.BotAPI {
	return app.bots
}

// ==================================================
// TRIAL CLEANUP TICKER
// ==================================================

func (app *BotApp) startTrialCleanupTicker(
	ctx context.Context,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-ticker.C:
				for botID := range app.bots {
					if err := app.SubscriptionService.CleanupExpiredTrials(ctx, botID); err != nil {
						app.ErrorNotify.Notify(
							ctx,
							botID,
							err,
							"Ошибка очистки истёкших trial-подписок",
						)
					}
				}
			}
		}
	}()
}

func (app *BotApp) SetAdminBotUsername(username string) {
	app.adminBotUsername = username
}
