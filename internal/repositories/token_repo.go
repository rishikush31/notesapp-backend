package repositories

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"notesapp-backend/internal/models"
)

var (
	ErrTokenNotFound = errors.New("refresh token not found")
)

// ------------------- INTERFACE -------------------
type TokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	GetByTokenID(ctx context.Context, id primitive.ObjectID) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id primitive.ObjectID) error
	RevokeAllForUser(ctx context.Context, userID primitive.ObjectID) error
}

// ------------------- STRUCT -------------------
type tokenRepo struct {
	collection *mongo.Collection
	infoLog    *log.Logger
	errorLog   *log.Logger
}

// ------------------- CONSTRUCTOR -------------------
func NewTokenRepository(db *mongo.Database, infoLog, errorLog *log.Logger) TokenRepository {
	tokenRepository := &tokenRepo{
		collection: db.Collection("refresh_tokens"),
		infoLog:    infoLog,
		errorLog:   errorLog,
	}

	tokenRepository.infoLog.Printf("Testing tokenRepository infoLogger")
	tokenRepository.errorLog.Printf("Testing tokenRepository errorLogger")

	return tokenRepository
}

// ------------------- CRUD METHODS -------------------

func (r *tokenRepo) Create(ctx context.Context, token *models.RefreshToken) error {
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, token)
	if err != nil {
		r.errorLog.Printf("token_repo.Create: %v", err)
		return err
	}
	return nil
}

func (r *tokenRepo) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{
		"token_hash": hash,
		"revoked":   false,
	}).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTokenNotFound
		}
		r.errorLog.Printf("token_repo.FindByHash: %v", err)
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepo) GetByTokenID(ctx context.Context, id primitive.ObjectID) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrTokenNotFound
		}
		r.errorLog.Printf("token_repo.GetByTokenID: %v", err)
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepo) Revoke(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"revoked": true}},
	)
	if err != nil {
		r.errorLog.Printf("token_repo.Revoke: %v", err)
		return err
	}
	return nil
}

func (r *tokenRepo) RevokeAllForUser(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"userId": userID},
		bson.M{"$set": bson.M{"revoked": true}},
	)
	if err != nil {
		r.errorLog.Printf("token_repo.RevokeAllForUser: %v", err)
		return err
	}
	return nil
}
