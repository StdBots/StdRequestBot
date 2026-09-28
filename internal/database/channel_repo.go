package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChannelRepo manages channel settings and counters
type ChannelRepo struct {
	channels *mongo.Collection
	audits   *mongo.Collection
}

// NewChannelRepo creates a new Channel repository
func NewChannelRepo(db *MongoDB) *ChannelRepo {
	return &ChannelRepo{
		channels: db.Channels,
		audits:   db.Audits,
	}
}

// UpsertChannel adds or updates channel profile
func (r *ChannelRepo) UpsertChannel(ch *ManagedChannel) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": ch.ChatID}
	update := bson.M{
		"$set": bson.M{
			"title":      ch.Title,
			"username":   ch.Username,
			"is_active":  true,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"_id":            ch.ChatID,
			"custom_welcome": "",
			"approved_count": 0,
			"added_at":       time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := r.channels.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetChannel fetches channel settings
func (r *ChannelRepo) GetChannel(chatID int64) (*ManagedChannel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var ch ManagedChannel
	err := r.channels.FindOne(ctx, bson.M{"_id": chatID}).Decode(&ch)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &ch, nil
}

// GetAllChannels returns all connected channels
func (r *ChannelRepo) GetAllChannels() ([]*ManagedChannel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "approved_count", Value: -1}})
	cursor, err := r.channels.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []*ManagedChannel
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// SetCustomWelcome updates custom welcome text for a channel
func (r *ChannelRepo) SetCustomWelcome(chatID int64, welcomeText string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.channels.UpdateOne(ctx,
		bson.M{"_id": chatID},
		bson.M{
			"$set": bson.M{
				"custom_welcome": welcomeText,
				"updated_at":     time.Now(),
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// IncrementApprovedCount increments approved member counter for a channel
func (r *ChannelRepo) IncrementApprovedCount(chatID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.channels.UpdateOne(ctx,
		bson.M{"_id": chatID},
		bson.M{"$inc": bson.M{"approved_count": 1}},
	)
	return err
}

// LogAudit saves individual join request audit
func (r *ChannelRepo) LogAudit(audit JoinAudit) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	audit.CreatedAt = time.Now()
	_, err := r.audits.InsertOne(ctx, audit)
	return err
}

// CountChannels returns total managed channels
func (r *ChannelRepo) CountChannels() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.channels.CountDocuments(ctx, bson.M{})
}

// CountTotalApproved aggregates total approved count across all channels
func (r *ChannelRepo) CountTotalApproved() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: "$approved_count"}}},
		}}},
	}

	cursor, err := r.channels.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Total int64 `bson:"total"`
	}
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		return 0, nil
	}
	return results[0].Total, nil
}
