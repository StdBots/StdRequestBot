package bot

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/StdBots/StdRequestBot/internal/analytics"
	"github.com/StdBots/StdRequestBot/internal/broadcast"
	"github.com/StdBots/StdRequestBot/internal/config"
	"github.com/StdBots/StdRequestBot/internal/credit"
	"github.com/StdBots/StdRequestBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot orchestrates Telegram updates and the DM-first join request pipeline
type Bot struct {
	api             *tgbotapi.BotAPI
	cfg             *config.Config
	db              *database.MongoDB
	channelRepo     *database.ChannelRepo
	userRepo        *database.UserRepo
	dedupeCache     *database.DedupeCache
	middleware      *Middleware
	handler         *Handler
	callbackHandler *CallbackHandler
	joinHandler     *JoinRequestHandler
	broadcastEngine *broadcast.Engine
	analyticsEngine *analytics.Engine
}

// New creates and wires up dependencies for the bot
func New(cfg *config.Config, db *database.MongoDB) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}

	if cfg.Env == "development" {
		api.Debug = true
	}

	channelRepo := database.NewChannelRepo(db)
	userRepo := database.NewUserRepo(db)
	dedupeCache := database.NewDedupeCache(180 * time.Second)
	middleware := NewMiddleware(cfg, userRepo, api)

	handler := NewHandler(api, channelRepo, userRepo, middleware)
	callbackHandler := NewCallbackHandler(api, channelRepo, userRepo, middleware, handler)
	joinHandler := NewJoinRequestHandler(api, cfg, channelRepo, userRepo, dedupeCache)
	broadcastEngine := broadcast.NewEngine(api, userRepo, cfg.OwnerID)
	analyticsEngine := analytics.NewEngine(api, channelRepo, userRepo, cfg.OwnerID)

	return &Bot{
		api:             api,
		cfg:             cfg,
		db:              db,
		channelRepo:     channelRepo,
		userRepo:        userRepo,
		dedupeCache:     dedupeCache,
		middleware:      middleware,
		handler:         handler,
		callbackHandler: callbackHandler,
		joinHandler:     joinHandler,
		broadcastEngine: broadcastEngine,
		analyticsEngine: analyticsEngine,
	}, nil
}

// Start begins polling for updates
func (b *Bot) Start(ctx context.Context) error {
	log.Printf("🤖 StdRequestBot authorized on account @%s (ID: %d)", b.api.Self.UserName, b.api.Self.ID)

	// Verify Credit Integrity
	intact, tampered := credit.VerifyIntegrity()
	if !intact {
		log.Printf("[SECURITY WARNING] Integrity violation detected: %v", tampered)
	}
	credit.ReportForkStatus(b.api.Self.UserName, intact)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Shutting down bot update loop...")
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			go b.processUpdate(update)
		}
	}
}

func (b *Bot) processUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC RECOVERED] in update %d: %v", update.UpdateID, r)
		}
	}()

	// 1. CRITICAL: Handle Incoming ChatJoinRequest updates!
	if update.ChatJoinRequest != nil {
		b.joinHandler.HandleJoinRequest(update.ChatJoinRequest)
		return
	}

	// 2. Handle Callback Queries
	if update.CallbackQuery != nil {
		b.callbackHandler.Handle(update.CallbackQuery)
		return
	}

	// 3. Handle Messages
	if update.Message != nil {
		msg := update.Message

		if msg.IsCommand() {
			cmd := strings.ToLower(msg.Command())
			switch cmd {
			case "start":
				b.handler.HandleStart(msg)
			case "channels", "mychannels":
				b.handler.HandleChannels(msg)
			case "setwelcome":
				b.handler.HandleSetWelcome(msg)
			case "delwelcome":
				b.handler.HandleDelWelcome(msg)
			case "help":
				b.handler.HandleHelp(msg)
			case "stats", "std":
				b.analyticsEngine.HandleStats(msg)
			case "broadcast":
				b.broadcastEngine.HandleBroadcast(msg)
			case "cancel":
				b.broadcastEngine.CancelBroadcast(msg.From.ID)
				reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Operation cancelled.")
				_, _ = b.api.Send(reply)
			default:
				reply := tgbotapi.NewMessage(msg.Chat.ID, "❓ Unknown command. Send /help to view available commands.")
				_, _ = b.api.Send(reply)
			}
			return
		}

		// Check if broadcast is waiting for content
		if b.broadcastEngine.IsAwaitingBroadcast(msg.From.ID) {
			b.broadcastEngine.ExecuteBroadcast(msg)
			return
		}
	}
}
