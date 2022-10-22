package app

import (
	"s3m/infrastructure/aws"
	"s3m/infrastructure/tgbot"
	"s3m/internal/config"
	"s3m/internal/usecase"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CmdClear = "clear"
	CmdCopy  = "copy"
)

type App struct {
	// logger  log.Logger
	usClear usecase.ClearUseCase
	cmd     string
}

func New(cfg *config.Config) (App, error) {

	var app App
	s3cli, err := aws.NewS3Client(cfg.AWS)
	if err != nil {
		return app, err
	}

	var tgBot *tgbotapi.BotAPI
	tgBot, err = tgbot.NewBot(cfg.TG.BotToken, true)
	if err != nil {
		return app, err
	}

	app.usClear = usecase.NewClearUseCase(
		s3cli,
		cfg.FolderSkip,
		cfg.AWS.Bucket,
		cfg.AWS.Prefix,
		cfg.AWS.Delimeter,
		cfg.Source,
		tgBot,
		cfg.TG.ChatID,
	)
	app.cmd = cfg.Subcommand

	return app, nil
}

func (a App) Run() error {
	switch a.cmd {
	case CmdClear:
		return a.usClear.Clear()
	case CmdCopy:
		return a.usClear.Copy()
	}
	return nil
}
