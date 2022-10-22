package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"s3m/infrastructure/aws"
	"s3m/internal/app"
	"s3m/internal/config"
)

var accessKey string
var secretKey string
var url string
var bucket string
var prefix string
var region string
var partition string
var folderSkip int

var tgToken string
var tgChat int

var source string

var defaultDelimeter string
var defaultChatID int

const (
	defaultUrl         = "https://storage.yandexcloud.net"
	defaultRegion      = "ru-central1"
	defaultPartition   = "yc"
	defaultFoldersSkip = 7
)

func main() {
	fs := &flag.FlagSet{}
	fs.StringVar(&accessKey, "access-key", "", "s3 login")
	fs.StringVar(&secretKey, "secret-key", "", "s3 secret")
	fs.StringVar(&url, "url", defaultUrl, "s3 url")
	fs.StringVar(&bucket, "bucket", "", "s3 bucket")
	fs.StringVar(&prefix, "prefix", "", "path prefix")
	fs.StringVar(&region, "region", defaultRegion, "s3 region")
	fs.StringVar(&partition, "partition", defaultPartition, "s3 partition")
	fs.StringVar(&tgToken, "tg-token", "", "tg bot token")
	fs.IntVar(&tgChat, "tg-chat", defaultChatID, "tg chat id")

	subcommand := os.Args[1]
	switch subcommand {
	case app.CmdClear:
		fs.IntVar(&folderSkip, "skip", defaultFoldersSkip, "skip folders")
	case app.CmdCopy:
		fs.StringVar(&source, "source", "", "source path for copy")
	default:
		fmt.Println("Need subcommand (copy, clear)", flag.Arg(0))
		os.Exit(1)
	}

	fs.Parse(os.Args[2:])

	// res := flag.NArg()
	// fmt.Println(res)
	if fs.NArg() != 0 {
		fmt.Println("All arguments must have -key before them. Incorrect argument:", fs.Arg(0))
		os.Exit(1)
	}

	if accessKey == "" {
		fmt.Println("-access-key is mandatory")
		os.Exit(1)
	}

	if secretKey == "" {
		fmt.Println("-secret-key is mandatory")
		os.Exit(1)
	}

	if bucket == "" {
		fmt.Println("-bucket is mandatory")
		os.Exit(1)
	}

	if prefix == "" {
		fmt.Println("-prefix is mandatory")
		os.Exit(1)
	}

	if subcommand == app.CmdCopy && source == "" {
		fmt.Println("-source is mandatory")
		os.Exit(1)
	}

	cfg := config.Config{
		AWS: aws.Config{
			AccessKey:     accessKey,
			SecretKey:     secretKey,
			Bucket:        bucket,
			Prefix:        prefix,
			Delimeter:     defaultDelimeter,
			PartitionID:   partition,
			URL:           url,
			SigningRegion: region,
		},
		Subcommand: subcommand,
		FolderSkip: folderSkip,
		Source:     source,
		TG: struct {
			BotToken string
			ChatID   int64
		}{
			BotToken: tgToken,
			ChatID:   int64(tgChat),
		},
	}
	a, err := app.New(&cfg)
	if err != nil {
		log.Fatalln(err)
	}
	err = a.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
