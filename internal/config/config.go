package config

import (
	"fmt"
	"os"
	"s3m/infrastructure/aws"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AWS        aws.Config
	Subcommand string
	FolderSkip int
	Source     string
	TG         struct {
		BotToken string
		ChatID   int64
	}
}

var instance *Config

func LoadConfig() *Config {
	instance = &Config{}
	if err := cleanenv.ReadEnv(instance); err != nil {
		helpText := "s3m - s3 manager"
		help, _ := cleanenv.GetDescription(instance, &helpText)
		fmt.Println(help)
		fmt.Println(err)
		os.Exit(1)
	}
	return instance
}

func GetConfig() *Config {
	return instance
}
