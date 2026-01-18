package repositories

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"notesapp-backend/internal/models"
)

// -------------------- INTERFACE --------------------

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByGoogleSub(ctx context.Context, googleSub string) (*models.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
}

// -------------------- IMPLEMENTATION --------------------

type userRepo struct {
	collection *mongo.Collection
	infoLog    *log.Logger
	errorLog   *log.Logger
}

func NewUserRepository(
	db *mongo.Database,
	infoLog, errorLog *log.Logger,
) UserRepository {
	userRepository := &userRepo{
		collection: db.Collection("users"), // internally picks collection
		infoLog:    infoLog,
		errorLog:   errorLog,
	}

	userRepository.infoLog.Printf("Testing userRepository infoLogger")
	userRepository.errorLog.Printf("Testing userRepository errorLogger")

	return userRepository
}

// Create new user
func (r *userRepo) Create(ctx context.Context, user *models.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		r.errorLog.Printf("userRepo.Create: %v", err)
		return err
	}
	r.infoLog.Printf("userRepo.Create: user %s created", user.ID.Hex())
	return nil
}

// Find user by email
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		r.errorLog.Printf("userRepo.FindByEmail: %v", err)
		return nil, err
	}
	return &user, nil
}

// Find user by Google Sub
func (r *userRepo) FindByGoogleSub(ctx context.Context, googleSub string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"googleSub": googleSub}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		r.errorLog.Printf("userRepo.FindByGoogleSub: %v", err)
		return nil, err
	}
	return &user, nil
}

// Find user by ID
func (r *userRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		r.errorLog.Printf("userRepo.FindByID: %v", err)
		return nil, err
	}
	return &user, nil
}
