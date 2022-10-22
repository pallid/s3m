package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func NewBot(token string, debug bool) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	bot.Debug = debug
	return bot, nil
}

func NewMessage(chatID int64, text string) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(chatID, text)
}

func Send(bot *tgbotapi.BotAPI, msg tgbotapi.Chattable) (tgbotapi.Message, error) {
	return bot.Send(msg)
}
