package util

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/tebeka/selenium"
)

type ChatInfo struct {
	ChatID int64
	MessageID int
}

// Driver struct
// Selenium 드라이버 설정 관련 구조체
type Driver struct {
	Driver selenium.WebDriver
	Service *selenium.Service
	Bot *tgbotapi.BotAPI
	ChannelId int64
	BotMessageIDs []ChatInfo
}

// Account struct
// 네이버 ID, PW 구조체
type Account struct {
	ID string `json:"naverID"`
	PW string `json:"naverPW"`
}

type BotInfo struct {
	TOKEN string `json:"BOT_TOKEN"`
	ChannelID string `json:"CHANNEL_ID"`
}

type ArticleID struct {
	id int
	link string
}

// ControlData struct
// 카페 컨트롤을 위한 구조체
type ControlData struct {
	D *Driver
	articleUrl chan []ArticleID
}