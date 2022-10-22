package usecase

import (
	"context"
	"fmt"
	"log"
	"s3m/infrastructure/aws"
	"s3m/infrastructure/tgbot"
	"s3m/internal/entity"
	"sort"
	"strings"

	apkg "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	TEMP_DELETE = "УДАЛЕН %s измененный %s"
	TEMP_SKIP   = "ПРОПУЩЕН %s измененный %s"
	layout      = "02.01.2006 15:04:05"
)

type ClearUseCase struct {
	s3cli       *s3.Client
	s3bucket    string
	s3prefix    string
	s3delimeter string
	folderSkip  int
	tgBot       *tgbotapi.BotAPI
	tgChatID    int64
}

func NewClearUseCase(client *s3.Client, folderSkip int, s3bucket, s3prefix, s3delimeter string, tgBot *tgbotapi.BotAPI, tgChatID int64) ClearUseCase {
	return ClearUseCase{
		s3cli:       client,
		s3bucket:    s3bucket,
		s3prefix:    s3prefix,
		s3delimeter: s3delimeter,
		folderSkip:  folderSkip,
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
		msg := tgbot.NewMessage(uc.tgChatID, generateClearMessages(messages))
		_, err := uc.tgBot.Send(msg)
		if err != nil {
			log.Println(err)
		}
	}

	return nil
}

func generateClearMessages(messages []string) string {
	text := "Результаты очистки старых каталогов: \n"
	for i := range messages {
		text = text + fmt.Sprintln(messages[i])
	}
	return text
}
