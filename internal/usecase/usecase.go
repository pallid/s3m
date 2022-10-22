package usecase

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"s3m/infrastructure/aws"
	"s3m/infrastructure/tgbot"
	"s3m/internal/entity"
	"sort"
	"strings"
	"time"

	apkg "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	TEMP_DELETE     = "УДАЛЕН %s измененный %s"
	TEMP_SKIP       = "ПРОПУЩЕН %s измененный %s"
	layout          = "02.01.2006 15:04:05"
	layoutNowFolder = "02-01-2006"
)

type ClearUseCase struct {
	s3cli       *s3.Client
	s3bucket    string
	s3prefix    string
	s3delimeter string
	folderSkip  int
	source      string
	tgBot       *tgbotapi.BotAPI
	tgChatID    int64
}

func NewClearUseCase(client *s3.Client, folderSkip int, s3bucket, s3prefix, s3delimeter string, source string, tgBot *tgbotapi.BotAPI, tgChatID int64) ClearUseCase {
	return ClearUseCase{
		s3cli:       client,
		s3bucket:    s3bucket,
		s3prefix:    s3prefix,
		s3delimeter: s3delimeter,
		folderSkip:  folderSkip,
		source:      source,
		tgBot:       tgBot,
		tgChatID:    tgChatID,
	}
}

func (uc ClearUseCase) Clear() error {

	paginator := aws.GetListObjectsPaginator(uc.s3cli, uc.s3bucket, uc.s3prefix, uc.s3delimeter)

	folders := map[string]*entity.Folder{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return err
		}
		for _, object := range page.Contents {
			// Do whatever you need with each object "obj"
			path := apkg.ToString(object.Key)
			tokens := strings.Split(strings.Trim(strings.Replace(path, uc.s3prefix, "", -1), "/"), "/")
			f := tokens[0]
			if fol, ok := folders[f]; !ok {
				fol = &entity.Folder{
					Name:     f,
					Modified: *object.LastModified,
				}
				fol.Deleted = append(fol.Deleted, types.ObjectIdentifier{Key: object.Key})
				folders[f] = fol
			} else {
				fol.Deleted = append(fol.Deleted, types.ObjectIdentifier{Key: object.Key})
			}
		}
	}

	var slice entity.TimeSliceFolder
	for _, v := range folders {
		slice = append(slice, *v)
	}
	sort.Sort(slice)

	var messages []string
	for i, item := range slice {
		// считаем от нуля, по этому -1
		if i > uc.folderSkip-1 {
			delMessage := fmt.Sprintf(TEMP_DELETE, item.Name, item.Modified.Format(layout))
			messages = append(messages, delMessage)
			log.Println(delMessage)
			err := aws.DeleteObjects(uc.s3cli, uc.s3bucket, item.Deleted)
			if err != nil {
				log.Println(err)
			}
		} else {
			skipMessage := fmt.Sprintf(TEMP_SKIP, item.Name, item.Modified.Format(layout))
			messages = append(messages, skipMessage)
			log.Println(skipMessage)
		}

	}

	if uc.tgBot != nil && len(messages) != 0 {
		msg := tgbot.NewMessage(uc.tgChatID, generateClearMessage(messages))
		_, err := uc.tgBot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func (uc ClearUseCase) Copy() error {

	nowFolder := time.Now().Format(layoutNowFolder)
	walker := make(fileWalk)
	go func() {
		// Gather the files to upload by walking the path recursively
		if err := filepath.Walk(uc.source, walker.Walk); err != nil {
			log.Fatalln("Walk failed:", err)
		}
		close(walker)
	}()

	for path := range walker {
		rel, err := filepath.Rel(uc.source, path)
		if err != nil {
			log.Fatalln("Unable to get relative path:", path, err)
		}
		file, err := os.Open(path)
		if err != nil {
			log.Println("Failed opening file", path, err)
			continue
		}
		defer file.Close()
		filePath := filepath.Join(uc.s3prefix, nowFolder, rel)
		filePath = strings.ReplaceAll(filePath, "\\", "/")
		err = aws.PutObject(uc.s3cli, uc.s3bucket, filePath, file)
		if err != nil {
			log.Fatalln("Failed to upload", path, err)
		}
		log.Println("Uploaded", path, filePath)
	}

	var messages []string
	messages = append(messages, "Добавлен новый архив "+nowFolder)
	if uc.tgBot != nil && len(messages) != 0 {
		msg := tgbot.NewMessage(uc.tgChatID, generateCopyMessage(messages))
		_, err := uc.tgBot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

type fileWalk chan string

func (f fileWalk) Walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		f <- path
	}
	return nil
}

func generateClearMessage(messages []string) string {
	text := "Результаты очистки старых каталогов: \n"
	for i := range messages {
		text = text + fmt.Sprintln(messages[i])
	}
	return text
}

func generateCopyMessage(messages []string) string {
	text := "Копирование в S3: \n"
	for i := range messages {
		text = text + fmt.Sprintln(messages[i])
	}
	return text
}
