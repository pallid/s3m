package app

import (
	"s3m/infrastructure/aws"
	"s3m/internal/config"
	"s3m/internal/usecase"
)

type App struct {
	// logger  log.Logger
	usClear usecase.ClearUseCase
}

func New(cfg *config.Config) (App, error) {

	var app App
	s3cli, err := aws.NewS3Client(cfg.AWS)
	if err != nil {
		return app, err
	}
	app.usClear = usecase.NewClearUseCase(
		s3cli,
		cfg.FolderSkip,
		cfg.AWS.Bucket,
		cfg.AWS.Prefix,
		cfg.AWS.Delimeter,
	)

	return app, nil
}

func (a App) Run() error {
	return a.usClear.Clear()
}
