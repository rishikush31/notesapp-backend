package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// -------------------- CLIENT --------------------

func NewMongoClient(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

// -------------------- INDEXES --------------------

func EnsureIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ================= USERS =================

	// Unique email
	_, err := db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"email", 1}},
		Options: options.Index().
			SetUnique(true).
			SetSparse(true),
	})
	if err != nil {
		return err
	}

	// Unique Google Sub
	_, err = db.Collection("users").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"googleSub", 1}},
		Options: options.Index().
			SetUnique(true).
			SetSparse(true),
	})
	if err != nil {
		return err
	}

	// ================= REFRESH TOKENS =================

	// Token hash lookup
	_, err = db.Collection("refresh_tokens").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"tokenHash", 1}},
	})
	if err != nil {
		return err
	}

	// TTL expiry
	_, err = db.Collection("refresh_tokens").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"expiresAt", 1}},
		Options: options.Index().
			SetExpireAfterSeconds(0),
	})
	if err != nil {
		return err
	}

	// ================= NOTES =================

	// List notes by user
	_, err = db.Collection("notes").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{"userId", 1}},
	})
	if err != nil {
		return err
	}

	// Ownership check
	_, err = db.Collection("notes").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{"_id", 1},
			{"userId", 1},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
