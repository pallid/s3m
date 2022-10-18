package usecase

import (
	"context"
	"log"
	"s3m/infrastructure/aws"
	"s3m/internal/entity"
	"sort"
	"strings"

	apkg "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type ClearUseCase struct {
	s3cli       *s3.Client
	s3bucket    string
	s3prefix    string
	s3delimeter string
	folderSkip  int
}

func NewClearUseCase(client *s3.Client, folderSkip int, s3bucket, s3prefix, s3delimeter string) ClearUseCase {
	return ClearUseCase{
		s3cli:       client,
		s3bucket:    s3bucket,
		s3prefix:    s3prefix,
		s3delimeter: s3delimeter,
		folderSkip:  folderSkip,
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

	for i, item := range slice {
		// считаем от нуля, по этому -1
		if i > uc.folderSkip-1 {
			log.Printf("WILL BE DELETE %s mod %v", item.Name, item.Modified)
			// log.Printf("count %v", len(item.Deleted))
			err := aws.DeleteObjects(uc.s3cli, uc.s3bucket, item.Deleted)
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Printf("SKIP %s mod %v", item.Name, item.Modified)
		}

	}

	return nil
}
