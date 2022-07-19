package client

import (
	"sync"

	"zheleznovux.com/modbus-console/cmd/serverStorage/tag"
)

type ClientInterface interface {
	Start(stop chan struct{}, wg *sync.WaitGroup)
	// SetType()
	Name() string
	Tags() []tag.TagInterface
	TagById(id int) (tag.TagInterface, error)
	TagByName(name string) (tag.TagInterface, error)
	SetTag(tag.TagInterface) error
	SetTags(tags []tag.TagInterface) error
}
