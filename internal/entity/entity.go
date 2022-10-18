package entity

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Folder struct {
	Name     string
	Modified time.Time
	Deleted  []types.ObjectIdentifier
}

type TimeSliceFolder []Folder

func (p TimeSliceFolder) Len() int {
	return len(p)
}

func (p TimeSliceFolder) Less(i, j int) bool {
	return p[i].Modified.After(p[j].Modified)
}

func (p TimeSliceFolder) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
