package analytics

import (
	"fmt"
	"runtime"
	"time"

	"github.com/StdBots/StdRequestBot/internal/credit"
	"github.com/StdBots/StdRequestBot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var startTime = time.Now()

// Engine renders system health and bot statistics
type Engine struct {
	api         *tgbotapi.BotAPI
	channelRepo *database.ChannelRepo
	userRepo    *database.UserRepo
	ownerID     int64
}

// NewEngine creates a new analytics engine
func NewEngine(api *tgbotapi.BotAPI, channelRepo *database.ChannelRepo, userRepo *database.UserRepo, ownerID int64) *Engine {
	return &Engine{
		api:         api,
		channelRepo: channelRepo,
		userRepo:    userRepo,
		ownerID:     ownerID,
	}
}

// HandleStats renders the system metrics and bot usage
func (e *Engine) HandleStats(msg *tgbotapi.Message) {
	if msg.From.ID != e.ownerID {
		reply := tgbotapi.NewMessage(msg.Chat.ID, "⛔ Unauthorized: Admin statistics only.")
		_, _ = e.api.Send(reply)
		return
	}

	totalUsers, _ := e.userRepo.CountUsers()
	totalChannels, _ := e.channelRepo.CountChannels()
	totalApproved, _ := e.channelRepo.CountTotalApproved()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime).Round(time.Second)

	statsText := fmt.Sprintf(
		"📈 <b>StdRequestBot Statistics & Performance:</b>\n\n"+
			"👥 <b>Total Captured Leads:</b> %d\n"+
			"🏢 <b>Managed Channels/Groups:</b> %d\n"+
			"✅ <b>Total Approved Requests:</b> %d\n"+
			"⏱️ <b>Uptime:</b> %s\n\n"+
			"⚙️ <b>System Metrics:</b>\n"+
			"• <b>Go Version:</b> %s\n"+
			"• <b>Goroutines:</b> %d\n"+
			"• <b>Memory Alloc:</b> %.2f MB\n"+
			"• <b>Memory Sys:</b> %.2f MB\n"+
			"• <b>Garbage Collections:</b> %d\n\n"+
			"%s",
		totalUsers,
		totalChannels,
		totalApproved,
		uptime,
		runtime.Version(),
		runtime.NumGoroutine(),
		float64(m.Alloc)/(1024*1024),
		float64(m.Sys)/(1024*1024),
		m.NumGC,
		credit.GetFooter(),
	)

	reply := tgbotapi.NewMessage(msg.Chat.ID, credit.GetWatermarked(statsText))
	reply.ParseMode = "HTML"

	_, _ = e.api.Send(reply)
}
