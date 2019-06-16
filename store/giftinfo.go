package store

import (
	"time"
)

// GiftInfo struct
type GiftInfo struct {
	ID       string    `bson:"id"`
	Name     string    `bson:"name"`
	Price    int32     `bson:"price"`
	Type     string    `bson:"type"`
	TypeName string    `bson:"type_name"`
	Date     string    `bson:"date"`
	CreateAt time.Time `bson:"create_at"`
}

// GetUniqueIndex fn
func (gl *GiftInfo) GetUniqueIndex() [][]string {
	return [][]string{{"id", "date"}}
}
