package bot

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/StdBots/StdRequestBot/internal/credit"
	"github.com/StdBots/StdRequestBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler processes bot commands
type Handler struct {
	bot         *tgbotapi.BotAPI
	channelRepo *database.ChannelRepo
	userRepo    *database.UserRepo
	middleware  *Middleware
}

// NewHandler creates command handler
func NewHandler(bot *tgbotapi.BotAPI, channelRepo *database.ChannelRepo, userRepo *database.UserRepo, middleware *Middleware) *Handler {
	return &Handler{
		bot:         bot,
		channelRepo: channelRepo,
		userRepo:    userRepo,
		middleware:  middleware,
	}
}

// HandleStart handles /start command
func (h *Handler) HandleStart(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	isMember, _ := h.middleware.CheckForceSub(msg.From.ID)
	if !isMember {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ <b>Channel Membership Required</b>\n\nPlease join our updates channel before accessing bot features:")
		reply.ParseMode = "HTML"
		reply.ReplyMarkup = h.middleware.GetForceSubMarkup()
		_, _ = h.bot.Send(reply)
		return
	}

	welcomeText := fmt.Sprintf(
		"🚀 <b>Welcome to StdRequestBot!</b>\n\n"+
			"The most advanced Telegram <b>Join Request Automated Approver</b> & Lead Generator.\n\n"+
			"✨ <b>Features:</b>\n"+
			"• <b>⚡ Instant Approval:</b> Automatically approves incoming join requests in 0.1s\n"+
			"• <b>📬 DM-First Protocol:</b> Sends customized welcome message directly to user's PM\n"+
			"• <b>🏢 Multi-Channel:</b> Manage unlimited channels & groups from one bot\n"+
			"• <b>👥 Lead Generation:</b> Automatically saves every applicant into MongoDB\n"+
			"• <b>📢 Mass Broadcast:</b> Broadcast updates to thousands of captured users\n\n"+
			"⚡ <i>Engineered by STD DEEPANSHU (%s)</i>\n%s",
		credit.Domain,
		credit.GetFooter(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.GetWatermarked(welcomeText))
	reply.ParseMode = "HTML"

	botUsername := h.bot.Self.UserName
	btnAddChannel := tgbotapi.NewInlineKeyboardButtonURL("➕ Add to Channel", fmt.Sprintf("https://t.me/%s?startchannel=true&admin=invite_users", botUsername))
	btnAddGroup := tgbotapi.NewInlineKeyboardButtonURL("➕ Add to Group", fmt.Sprintf("https://t.me/%s?startgroup=true&admin=invite_users", botUsername))
	btnChannels := tgbotapi.NewInlineKeyboardButtonData("🏢 Connected Channels", "channels:list")
	btnDev := tgbotapi.NewInlineKeyboardButtonURL("👨‍💻 Developer", "https://"+credit.Domain)

	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnAddChannel, btnAddGroup),
		tgbotapi.NewInlineKeyboardRow(btnChannels, btnDev),
	)

	_, _ = h.bot.Send(reply)
}

// HandleChannels displays connected channels dashboard
func (h *Handler) HandleChannels(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	if !h.middleware.IsOwner(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ Channel management is restricted to administrators.")
		_, _ = h.bot.Send(reply)
		return
	}

	channels, err := h.channelRepo.GetAllChannels()
	if err != nil || len(channels) == 0 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "ℹ️ <b>No Channels Connected Yet!</b>\n\nTo connect a channel:\n1. Add this bot to your channel or group as an <b>Admin</b>.\n2. Ensure the bot has <b>'Invite Users via Link'</b> admin permission.\n3. Turn on <b>'Request to Join'</b> in your invite link settings!\n\nThe bot will automatically register the channel upon the first join request.")
		reply.ParseMode = "HTML"
		_, _ = h.bot.Send(reply)
		return
	}

	var sb strings.Builder
	sb.WriteString("🏢 <b>Connected Channels & Groups:</b>\n\n")

	for idx, ch := range channels {
		status := "🟢 Active"
		if !ch.IsActive {
			status = "🔴 Inactive"
		}
		sb.WriteString(fmt.Sprintf("%d. <b>%s</b>\n", idx+1, ch.Title))
		sb.WriteString(fmt.Sprintf("   🆔 <code>%d</code> | %s\n", ch.ChatID, status))
		sb.WriteString(fmt.Sprintf("   👥 Approved Members: <b>%d</b>\n", ch.ApprovedCount))

		if ch.CustomWelcome != "" {
			sb.WriteString("   📝 Custom Welcome: <i>Configured ✅</i>\n")
		} else {
			sb.WriteString("   📝 Custom Welcome: <i>Default Funnel</i>\n")
		}
		sb.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
	}

	sb.WriteString("\n💡 <i>To set a custom welcome message:</i>\n<code>/setwelcome &lt;chat_id&gt; Your message here</code>\n")
	sb.WriteString(credit.GetFooter())

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.GetWatermarked(sb.String()))
	reply.ParseMode = "HTML"

	btnRefresh := tgbotapi.NewInlineKeyboardButtonData("🔄 Refresh Channels", "channels:list")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btnRefresh))

	_, _ = h.bot.Send(reply)
}

// HandleSetWelcome sets a custom welcome message for a channel
func (h *Handler) HandleSetWelcome(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	if !h.middleware.IsOwner(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ Admin access only.")
		_, _ = h.bot.Send(reply)
		return
	}

	args := strings.TrimSpace(msg.CommandArguments())
	parts := strings.SplitN(args, " ", 2)

	if len(parts) < 2 {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ <b>Usage:</b>\n<code>/setwelcome &lt;chat_id&gt; &lt;your_welcome_message&gt;</code>\n\n<b>Supported Placeholders:</b>\n• <code>{name}</code> — User's first name\n• <code>{chat_title}</code> — Channel or group title\n• <code>{user_id}</code> — User numeric ID")
		reply.ParseMode = "HTML"
		_, _ = h.bot.Send(reply)
		return
	}

	chatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Invalid chat_id. It must be a numeric ID (e.g. -1001234567890).")
		_, _ = h.bot.Send(reply)
		return
	}

	welcomeText := parts[1]
	err = h.channelRepo.SetCustomWelcome(chatID, welcomeText)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "❌ Failed to save custom welcome message to database.")
		_, _ = h.bot.Send(reply)
		return
	}

	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("✅ <b>Custom Welcome Saved!</b>\n\n<b>Chat ID:</b> <code>%d</code>\n<b>Message Preview:</b>\n%s", chatID, welcomeText))
	reply.ParseMode = "HTML"
	_, _ = h.bot.Send(reply)
}

// HandleDelWelcome resets a channel welcome message to default
func (h *Handler) HandleDelWelcome(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	if !h.middleware.IsOwner(msg.From.ID) {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ Admin access only.")
		_, _ = h.bot.Send(reply)
		return
	}

	args := strings.TrimSpace(msg.CommandArguments())
	chatID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⚠️ <b>Usage:</b>\n<code>/delwelcome &lt;chat_id&gt;</code>")
		reply.ParseMode = "HTML"
		_, _ = h.bot.Send(reply)
		return
	}

	_ = h.channelRepo.SetCustomWelcome(chatID, "")

	reply := tgbotapi.NewMessage(msg.Chat.ID, fmt.Sprintf("✅ Custom welcome removed for <code>%d</code>. Channel will now use the default funnel template.", chatID))
	reply.ParseMode = "HTML"
	_, _ = h.bot.Send(reply)
}

// HandleHelp displays guide and permissions
func (h *Handler) HandleHelp(msg *tgbotapi.Message) {
	h.middleware.TrackUser(msg.From)

	helpText := fmt.Sprintf(
		"📖 <b>StdRequestBot Setup & Usage Guide:</b>\n\n"+
			"<b>1. Add Bot as Admin:</b>\n"+
			"• Add <code>@%s</code> to your channel or group as an <b>Admin</b>.\n"+
			"• Give the bot the <b>'Invite Users via Link'</b> admin right (required to approve requests).\n\n"+
			"<b>2. Enable Join Requests:</b>\n"+
			"• Go to Channel Settings ➡️ Invite Links.\n"+
			"• Create a new link and enable <b>'Request Admin Approval'</b>.\n\n"+
			"<b>3. Auto-Approval:</b>\n"+
			"• Whenever a user clicks your link, the bot immediately sends a welcome DM and approves them!\n\n"+
			"<b>Commands:</b>\n"+
			"• /start — Launch bot and view status\n"+
			"• /channels — List connected channels and approved counts (Admin only)\n"+
			"• /setwelcome — Set custom welcome message per channel\n"+
			"• /delwelcome — Reset channel welcome message to default\n"+
			"• /stats — Global bot performance and metrics (Admin only)\n"+
			"• /broadcast — Mass broadcast to all captured users (Admin only)\n\n"+
			"%s",
		h.bot.Self.UserName,
		credit.GetFooter(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.GetWatermarked(helpText))
	reply.ParseMode = "HTML"
	_, _ = h.bot.Send(reply)
}
