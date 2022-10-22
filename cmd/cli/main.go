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

var defaultDelimeter string
var defaultChatID int

const (
	defaultUrl         = "https://storage.yandexcloud.net"
	defaultRegion      = "ru-central1"
	defaultPartition   = "yc"
	defaultFoldersSkip = 7
)

func main() {

	flag.StringVar(&accessKey, "access-key", "", "s3 login")
	flag.StringVar(&secretKey, "secret-key", "", "s3 secret")
	flag.StringVar(&url, "url", defaultUrl, "s3 url")
	flag.StringVar(&bucket, "bucket", "", "s3 bucket")
	flag.StringVar(&prefix, "prefix", "", "path prefix")
	flag.StringVar(&region, "region", defaultRegion, "s3 region")
	flag.StringVar(&partition, "partition", defaultPartition, "s3 partition")
	flag.IntVar(&folderSkip, "skip", defaultFoldersSkip, "skip folders")
	flag.StringVar(&tgToken, "tg-token", "", "tg bot token")
	flag.IntVar(&tgChat, "tg-chat", defaultChatID, "tg chat id")

	flag.Parse()

	if flag.NArg() != 0 {
		fmt.Println("All arguments must have -key before them. Incorrect argument:", flag.Arg(0))
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
		FolderSkip: folderSkip,
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
