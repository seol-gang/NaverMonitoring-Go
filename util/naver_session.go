package util

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// Enabled function
// Selenium 명시적 대기를 위한 함수
func Enabled(by, elementName string) func(selenium.WebDriver) (bool, error) {
	return func(wd selenium.WebDriver) (bool, error) {
		el, err := wd.FindElement(by, elementName)
		if err != nil {
			return false, nil
		}
		enabled, err := el.IsEnabled()
		if err != nil {
			return false, nil
		}

		if !enabled {
			return false, nil
		}

		return true, nil
	}
}

// GetAccountInfo fuction
// 네이버 ID, PW 가져오는 함수
func GetAccountInfo() (string, string) {
	jsonFile, err := os.Open("./account.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	jsonValue, _ := ioutil.ReadAll(jsonFile)

	var accountInfo Account
	json.Unmarshal(jsonValue, &accountInfo)
	return accountInfo.ID, accountInfo.PW
}

// SetTelegramBot function
// 텔레그램 봇 셋팅 함수
func SetTelegramBot() (*tgbotapi.BotAPI, int64) {
	jsonFile, err := os.Open("./bot_info.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	jsonValue, _ := ioutil.ReadAll(jsonFile)

	var botInfo BotInfo
	json.Unmarshal(jsonValue, &botInfo)

	bot, err := tgbotapi.NewBotAPI(botInfo.TOKEN)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	channelId, _ := strconv.ParseInt(botInfo.ChannelID, 10, 64)

	return bot, channelId
}

func (c *Driver) LoopUpdateMessage() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := c.Bot.GetUpdatesChan(u)
	if err != nil {
		fmt.Println(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.Text == "확인" {
			for _, msgInfo := range c.BotMessageIDs {
				deleteConfig := tgbotapi.DeleteMessageConfig{
					ChatID: msgInfo.ChatID,
					MessageID: msgInfo.MessageID,
				}
				_, err := c.Bot.DeleteMessage(deleteConfig)
				if err != nil {
					fmt.Println(err)
				}
			}
			c.BotMessageIDs = []ChatInfo{}
			updates.Clear()
		}
	}
}

// SendBotMessage function
// 텔레그램 봇 메세지 전송 함수
func (c *Driver) SendBotMessage(message string) {
	msg := tgbotapi.NewMessage(c.ChannelId, message)
	sendMsg, _ := c.Bot.Send(msg)
	msgInfo := ChatInfo{
		ChatID: sendMsg.Chat.ID,
		MessageID: sendMsg.MessageID,
	}
	c.BotMessageIDs = append(c.BotMessageIDs, msgInfo)
}

// RunSeleniumClient function
// Selenium 클라이언트 시작을 위한 함수
func RunSeleniumClient(port int) (selenium.WebDriver, *selenium.Service) {
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			//"--headless",
		},
	}

	caps.AddChrome(chromeCaps)

	service, err := selenium.NewChromeDriverService("./chromedriver", port)
	if err != nil {
		fmt.Println(err)
	}

	wd, err := selenium.NewRemote(caps, "")
	if err != nil {
		fmt.Println(err)
	}

	return wd, service
}

// LoginNaver function
func LoginNaver(driver selenium.WebDriver, id, pw string) error {
	script := `
	(function execute(){
		document.querySelector('#id').value = "` + id + `";
		document.querySelector('#pw').value = "` + pw + `";
	})();
	`
	driver.ExecuteScript(script, nil)
	if err := driver.Wait(Enabled(selenium.ByCSSSelector, "input.btn_global")); err != nil {
		return err
	}
	element, _ := driver.FindElement(selenium.ByCSSSelector, "input.btn_global")
	element.Click()
	return nil
}

func NewDriver(port int) *Driver {
	wd, service := RunSeleniumClient(port)
	id, pw := GetAccountInfo()
	bot, channelId := SetTelegramBot()
	wd.Get("https://nid.naver.com/nidlogin.login")
	if err := wd.Wait(Enabled(selenium.ByCSSSelector, `#log\.login`)); err != nil {
		fmt.Println(err)
		return nil
	}
	LoginNaver(wd, id, pw)
	if err := wd.Wait(Enabled(selenium.ByCSSSelector, "#footer > div > div.corp_area > address > a")); err != nil {
		fmt.Println(fmt.Sprintf("NewDriver Function error : %s", err))
		return nil
	}

	return &Driver{Driver: wd, Service: service, Bot: bot, ChannelId: channelId}
}