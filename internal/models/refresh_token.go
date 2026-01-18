package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RefreshToken represents a refresh token stored in DB.
// We store the HASH, not the raw token, for security.
type RefreshToken struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	UserID     primitive.ObjectID `bson:"user_id"`     // reference to User._id
	TokenHash  string             `bson:"token_hash"`  // hashed token
	DeviceInfo string             `bson:"device_info"` // optional device/session info
	ExpiresAt  time.Time          `bson:"expires_at"`  // token expiry
	Revoked    bool               `bson:"revoked"`     // manually revoked
	CreatedAt  time.Time          `bson:"created_at"`  // created timestamp
}
