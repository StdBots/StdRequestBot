package bot

import (
	"github.com/StdBots/StdRequestBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CallbackHandler processes inline button queries
type CallbackHandler struct {
	bot         *tgbotapi.BotAPI
	channelRepo *database.ChannelRepo
	userRepo    *database.UserRepo
	middleware  *Middleware
	handler     *Handler
}

// NewCallbackHandler initializes callback query handler
func NewCallbackHandler(bot *tgbotapi.BotAPI, channelRepo *database.ChannelRepo, userRepo *database.UserRepo, middleware *Middleware, handler *Handler) *CallbackHandler {
	return &CallbackHandler{
		bot:         bot,
		channelRepo: channelRepo,
		userRepo:    userRepo,
		middleware:  middleware,
		handler:     handler,
	}
}

// Handle processes incoming callback queries
func (c *CallbackHandler) Handle(query *tgbotapi.CallbackQuery) {
	c.middleware.TrackUser(query.From)
	data := query.Data

	switch data {
	case "channels:list":
		c.answer(query.ID, "🔄 Refreshing channels...", false)
		if query.Message != nil {
			c.handler.HandleChannels(query.Message)
		}
	default:
		c.answer(query.ID, "Action recognized.", false)
	}
}

func (c *CallbackHandler) answer(queryID string, text string, showAlert bool) {
	cb := tgbotapi.NewCallback(queryID, text)
	cb.ShowAlert = showAlert
	_, _ = c.bot.Request(cb)
}
