package store

import (
	"time"
)

// event const
const (
	RoomEventUserEnter = "UserEnter" // 进入房间
	RoomEventUserExit  = "UserExit"  // 离开房间
)

// Live struct
type Live struct {
	ID         string    `bson:"id"`
	Name       string    `bson:"name"`
	RoomID     string    `bson:"room_id"`
	RoomType   string    `bson:"room_type"`
	AnchorUID  string    `bson:"anchor_uid"`
	AnchorName string    `bson:"anchor_name"`
	StartedAt  time.Time `bson:"started_at"`
	CreatedAt  time.Time `bson:"created_at"`
}

func (l *Live) GetUniqueIndex() [][]string {
	return [][]string{{"id"}}
}

// Gift struct
type Gift struct {
	ID           string    `bson:"id"`
	Name         string    `bson:"name"`
	Num          int32     `bson:"num"`
	Price        int32     `bson:"price"`
	SenderUID    string    `bson:"sender_uid"`
	SenderName   string    `bson:"sender_name"`
	ReceiverUid  string    `bson:"receiver_uid"`
	ReceiverName string    `bson:"receiver_name"`
	RoomID       string    `bson:"room_id"`
	LiveID       string    `bson:"live_id"`
	CreatedAt    time.Time `bson:"created_at"`
}

// GetPartition fn
func (g *Gift) GetPartition() interface{} {
	return g.CreatedAt
}

type Message struct {
	SenderUid    string    `bson:"sender_uid"`
	SenderName   string    `bson:"sender_name"`
	ReceiverUid  string    `bson:"receiver_uid"`
	ReceiverName string    `bson:"receiver_name"`
	Content      string    `bson:"content"`
	RoomID       string    `bson:"room_id"`
	LiveID       string    `bson:"live_id"`
	CreatedAt    time.Time `bson:"created_at"`
}

func (m *Message) GetPartition() interface{} {
	return m.CreatedAt
}

// LiveEnd struct
type LiveEnd struct {
	LiveID    string    `bson:"live_id"`
	EndedAt   time.Time `bson:"ended_at"`
	Duration  int32     `bson:"duration"`
	CreatedAt time.Time `bson:"created_at"`
}

// GetUniqueIndex fn
func (l *LiveEnd) GetUniqueIndex() [][]string {
	return [][]string{{"live_id"}}
}

// RoomUserCount struct
type RoomUserCount struct {
	RoomID    string    `bson:"room_id"`
	LiveID    string    `bson:"live_id"`
	Count     int32     `bson:"count"`
	CreatedAt time.Time `bson:"created_at"`
}

// GetPartition fn
func (r RoomUserCount) GetPartition() interface{} {
	return r.CreatedAt
}

// RoomEvent struct
type RoomEvent struct {
	RoomID    string    `bson:"room_id"`
	LiveID    string    `bson:"live_id"`
	Type      string    `bson:"type"`
	Uid       string    `bson:"uid"`
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"created_at"`
}

// GetPartition fn
func (r RoomEvent) GetPartition() interface{} {
	return r.CreatedAt
}
