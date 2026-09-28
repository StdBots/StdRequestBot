package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/StdBots/StdRequestBot/internal/config"
	"github.com/StdBots/StdRequestBot/internal/credit"
	"github.com/StdBots/StdRequestBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// JoinRequestHandler manages the Critical DM-First Join Request Processing Pipeline
type JoinRequestHandler struct {
	bot         *tgbotapi.BotAPI
	cfg         *config.Config
	channelRepo *database.ChannelRepo
	userRepo    *database.UserRepo
	dedupeCache *database.DedupeCache
}

// NewJoinRequestHandler creates join request pipeline handler
func NewJoinRequestHandler(bot *tgbotapi.BotAPI, cfg *config.Config, channelRepo *database.ChannelRepo, userRepo *database.UserRepo, dedupeCache *database.DedupeCache) *JoinRequestHandler {
	return &JoinRequestHandler{
		bot:         bot,
		cfg:         cfg,
		channelRepo: channelRepo,
		userRepo:    userRepo,
		dedupeCache: dedupeCache,
	}
}

// HandleJoinRequest processes an incoming ChatJoinRequest update
func (h *JoinRequestHandler) HandleJoinRequest(req *tgbotapi.ChatJoinRequest) {
	chatID := req.Chat.ID
	user := req.From
	userID := user.ID
	chatTitle := req.Chat.Title
	if chatTitle == "" {
		chatTitle = "our chat"
	}

	// 0. Deduplication Guard: Ignore if processed within 180 seconds
	if h.dedupeCache.AlreadyHandled(chatID, userID) {
		log.Printf("[JOINREQ DEDUPE] Skipping duplicate request for user=%d in chat=%d", userID, chatID)
		return
	}

	log.Printf("»» [joinreq] Received join request: user=%d (%s) in chat=%d (%s)", userID, user.FirstName, chatID, chatTitle)

	// 1. Register/Update channel in MongoDB
	go func() {
		_ = h.channelRepo.UpsertChannel(&database.ManagedChannel{
			ChatID:   chatID,
			Title:    chatTitle,
			Username: req.Chat.UserName,
		})
	}()

	// 2. Prepare Welcome Message & Buttons
	welcomeText, markup := h.buildWelcomeMessage(req, chatTitle)

	// ── STEP 1: SEND DM FIRST WHILE REQUEST IS PENDING! ──
	// Telegram temporarily authorizes the bot to contact the user directly ONLY while the request is pending.
	dmMsg := tgbotapi.NewMessage(userID, credit.GetWatermarked(welcomeText))
	dmMsg.ParseMode = "HTML"
	dmMsg.ReplyMarkup = markup

	delivered := false
	_, err := h.bot.Send(dmMsg)
	if err == nil {
		delivered = true
		log.Printf("»» [joinreq] ✅ DM successfully delivered to user=%d (Chat: %s)", userID, chatTitle)
	} else {
		log.Printf("»» [joinreq] ⚠️ DM failed for user=%d: %v", userID, err)
	}

	// ── Optional Configurable Delay (Anti-Flood / Natural Approval) ──
	if h.cfg.ApprovalDelaySeconds > 0 {
		time.Sleep(time.Duration(h.cfg.ApprovalDelaySeconds) * time.Second)
	}

	// ── STEP 2: NOW APPROVE THE REQUEST ──
	approveConfig := tgbotapi.ApproveChatJoinRequestConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	}

	approved := false
	_, err = h.bot.Request(approveConfig)
	if err == nil {
		approved = true
		log.Printf("»» [joinreq] ✅ Approved user=%d in chat=%d (%s)", userID, chatID, chatTitle)
		_ = h.channelRepo.IncrementApprovedCount(chatID)
	} else {
		log.Printf("»» [joinreq] ❌ Failed to approve user=%d in chat=%d: %v", userID, chatID, err)
	}

	// ── STEP 3: LEAD REGISTRATION IN MONGODB ──
	// Save user to database for future broadcasts (marks dm_delivery status)
	go func() {
		_ = h.userRepo.RegisterLead(userID, user.UserName, user.FirstName, user.LastName, chatID, chatTitle, delivered)
		_ = h.channelRepo.LogAudit(database.JoinAudit{
			UserID:    userID,
			ChatID:    chatID,
			ChatTitle: chatTitle,
			DMSent:    delivered,
			Approved:  approved,
		})
	}()
}

// buildWelcomeMessage formats custom or default welcome caption with funnel buttons
func (h *JoinRequestHandler) buildWelcomeMessage(req *tgbotapi.ChatJoinRequest, chatTitle string) (string, tgbotapi.InlineKeyboardMarkup) {
	name := req.From.FirstName
	if name == "" {
		name = "there"
	}

	// Check if channel has custom welcome template in MongoDB
	customWelcome := ""
	ch, _ := h.channelRepo.GetChannel(req.Chat.ID)
	if ch != nil && ch.CustomWelcome != "" {
		customWelcome = ch.CustomWelcome
	}

	var text string
	if customWelcome != "" {
		text = strings.ReplaceAll(customWelcome, "{name}", name)
		text = strings.ReplaceAll(text, "{chat_title}", chatTitle)
		text = strings.ReplaceAll(text, "{user_id}", fmt.Sprintf("%d", req.From.ID))
	} else {
		// Default High-Conversion Funnel Template
		text = fmt.Sprintf(
			"🎉 <b>Welcome, %s!</b>\n\n"+
				"Your request to join <b>%s</b> has been approved! ✨\n\n"+
				"We're glad to have you with us. Enjoy your stay and explore our official tools below:\n\n"+
				"%s",
			name,
			chatTitle,
			credit.GetFooter(),
		)
	}

	// High-Conversion Funnel Buttons
	botUsername := h.bot.Self.UserName
	btnStart := tgbotapi.NewInlineKeyboardButtonURL("🤖 Open Bot", fmt.Sprintf("https://t.me/%s?start=joinreq", botUsername))
	btnAddGroup := tgbotapi.NewInlineKeyboardButtonURL("➕ Add Me To Your Group ➕", fmt.Sprintf("https://t.me/%s?startgroup=true", botUsername))
	btnChannel := tgbotapi.NewInlineKeyboardButtonURL("📢 Official Channel", "https://t.me/StdBots")

	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnStart),
		tgbotapi.NewInlineKeyboardRow(btnAddGroup),
		tgbotapi.NewInlineKeyboardRow(btnChannel),
	)

	return text, markup
}
