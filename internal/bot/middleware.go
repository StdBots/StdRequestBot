package bot

import (
	"fmt"
	"log"

	"github.com/StdBots/StdRequestBot/internal/config"
	"github.com/StdBots/StdRequestBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Middleware handles authentication and channel subscription
type Middleware struct {
	cfg      *config.Config
	userRepo *database.UserRepo
	bot      *tgbotapi.BotAPI
}

// NewMiddleware creates new Middleware
func NewMiddleware(cfg *config.Config, userRepo *database.UserRepo, bot *tgbotapi.BotAPI) *Middleware {
	return &Middleware{
		cfg:      cfg,
		userRepo: userRepo,
		bot:      bot,
	}
}

// TrackUser registers direct user interaction in MongoDB
func (m *Middleware) TrackUser(from *tgbotapi.User) {
	if from == nil {
		return
	}
	go func() {
		err := m.userRepo.RegisterLead(from.ID, from.UserName, from.FirstName, from.LastName, 0, "Direct Interaction", true)
		if err != nil {
			log.Printf("Error tracking user %d: %v", from.ID, err)
		}
	}()
}

// IsOwner checks if user is bot owner
func (m *Middleware) IsOwner(userID int64) bool {
	return userID == m.cfg.OwnerID
}

// CheckForceSub verifies channel subscription
func (m *Middleware) CheckForceSub(userID int64) (bool, error) {
	if m.cfg.ForceSubChannel == "" {
		return true, nil
	}

	chatConfig := tgbotapi.ChatInfoConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: "@" + m.cfg.ForceSubChannel,
			UserID:             userID,
		},
	}

	member, err := m.bot.GetChatMember(chatConfig)
	if err != nil {
		return true, nil
	}

	status := member.Status
	if status == "creator" || status == "administrator" || status == "member" || status == "restricted" {
		return true, nil
	}

	return false, nil
}

// GetForceSubMarkup returns force subscribe keyboard
func (m *Middleware) GetForceSubMarkup() tgbotapi.InlineKeyboardMarkup {
	channelURL := fmt.Sprintf("https://t.me/%s", m.cfg.ForceSubChannel)
	btnChannel := tgbotapi.NewInlineKeyboardButtonURL("📢 Join Updates Channel", channelURL)

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(btnChannel),
	)
}
