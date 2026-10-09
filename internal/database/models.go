package database

import "time"

// User represents a lead captured through join requests or direct bot interaction
type User struct {
	ID              int64     `bson:"_id" json:"id"`
	Username        string    `bson:"username" json:"username"`
	FirstName       string    `bson:"first_name" json:"first_name"`
	LastName        string    `bson:"last_name" json:"last_name"`
	SourceChatID    int64     `bson:"source_chat_id" json:"source_chat_id"`
	SourceChatTitle string    `bson:"source_chat_title" json:"source_chat_title"`
	DMDelivery      bool      `bson:"dm_delivery" json:"dm_delivery"`
	JoinedAt        time.Time `bson:"joined_at" json:"joined_at"`
	LastActive      time.Time `bson:"last_active" json:"last_active"`
}

// ManagedChannel represents a channel or group where the bot manages join requests
type ManagedChannel struct {
	ChatID        int64     `bson:"_id" json:"chat_id"`
	Title         string    `bson:"title" json:"title"`
	Username      string    `bson:"username" json:"username"`
	CustomWelcome string    `bson:"custom_welcome" json:"custom_welcome"`
	ApprovedCount int64     `bson:"approved_count" json:"approved_count"`
	IsActive      bool      `bson:"is_active" json:"is_active"`
	AddedAt       time.Time `bson:"added_at" json:"added_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

// JoinAudit stores individual join request processing logs
type JoinAudit struct {
	UserID    int64     `bson:"user_id" json:"user_id"`
	ChatID    int64     `bson:"chat_id" json:"chat_id"`
	ChatTitle string    `bson:"chat_title" json:"chat_title"`
	DMSent    bool      `bson:"dm_sent" json:"dm_sent"`
	Approved  bool      `bson:"approved" json:"approved"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
