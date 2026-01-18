package services

import (
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/api/idtoken"
	"github.com/golang-jwt/jwt/v5"

	"notesapp-backend/internal/config"
	"notesapp-backend/internal/models"
	"notesapp-backend/internal/repositories"
	"notesapp-backend/internal/utils"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("refresh token expired")
	ErrTokenRevoked       = errors.New("refresh token revoked")
	ErrInvalidToken       = errors.New("Invalid Token")
)

type AuthService struct {
	userRepo  repositories.UserRepository
	tokenRepo repositories.TokenRepository
	cfg       *config.Config
	infoLog   *log.Logger
	errorLog  *log.Logger
}

func NewAuthService(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	cfg *config.Config,
	infoLog *log.Logger,
	errorLog *log.Logger,
) *AuthService {
	authService := &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
		infoLog:   infoLog,
		errorLog:  errorLog,
	}
	
	authService.infoLog.Printf("Testing authService infoLogger")
	authService.errorLog.Printf("Testing authService errorLogger")

	return authService
}

// -------------------- REGISTER --------------------

func (s *AuthService) Register(
	ctx context.Context,
	name, email, password string,
) (*models.User, error) {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return nil, errors.New("email already exists")
	}
	if err != repositories.ErrUserNotFound {
		return nil, err
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		s.errorLog.Printf("auth.Register hash: %v", err)
		return nil, err
	}

	user := &models.User{
		ID:           primitive.NewObjectID(),
		Name:         name,
		Email:        email,
		PasswordHash: &hash,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// -------------------- LOGIN --------------------

func (s *AuthService) Login(
	ctx context.Context,
	email, password, deviceInfo string,
) (accessToken string, refreshToken string, err error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		return "", "", ErrInvalidCredentials
	}

	if !utils.ComparePassword(*user.PasswordHash, password) {
		return "", "", ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user.ID, deviceInfo)
}

// -------------------- GOOGLE LOGIN --------------------
var ErrInvalidGoogleToken = errors.New("invalid google token")

func (s *AuthService) GoogleLogin(
	ctx context.Context,
	googleIDToken string,
	deviceInfo string,
) (accessToken string, refreshToken string, err error) {

	// 1. Verify Google ID token
	payload, err := idtoken.Validate(ctx, googleIDToken, s.cfg.GoogleClientID)
	if err != nil {
		return "", "", ErrInvalidGoogleToken
	}

	// 2. Extract claims
	googleSub := payload.Subject

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return "", "", ErrInvalidGoogleToken
	}

	name, _ := payload.Claims["name"].(string)

	// 3. Find user by Google sub
	user, err := s.userRepo.FindByGoogleSub(ctx, googleSub)
	if err != nil {
		if err != repositories.ErrUserNotFound {
			return "", "", err
		}

		// 4. Create user if not exists
		user = &models.User{
			ID:        primitive.NewObjectID(),
			Name:      name,
			Email:     email,
			GoogleSub: &googleSub,
			CreatedAt: time.Now().UTC(),
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return "", "", err
		}
	}

	// 5. Issue app tokens
	return s.issueTokens(ctx, user.ID, deviceInfo)
}

// -------------------- REFRESH --------------------

func (s *AuthService) RefreshToken(
	ctx context.Context,
	rawRefreshToken string,
) (accessToken string, newRefreshToken string, err error) {
	// 1. Hash the incoming refresh token
	hash := utils.HashToken(rawRefreshToken)

	// 2. Find the stored token in DB
	storedToken, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil {
		return "", "", err // token not found
	}

	// 3. Check if token is revoked
	if storedToken.Revoked {
		return "", "", ErrTokenRevoked
	}

	// 4. Check if token is expired
	if time.Now().UTC().After(storedToken.ExpiresAt) {
		return "", "", ErrTokenExpired
	}

	// 5. Revoke old token
	if err := s.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return "", "", err
	}

	// 6. Issue new tokens (access + refresh)
	return s.issueTokens(ctx, storedToken.UserID, storedToken.DeviceInfo)
}

// -------------------- LOGOUT --------------------

func (s *AuthService) Logout(
	ctx context.Context,
	refreshToken string,
) error {
	hash := utils.HashToken(refreshToken)

	stored, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil {
		return err
	}

	return s.tokenRepo.Revoke(ctx, stored.ID)
}

// -------------------- INTERNAL --------------------

func (s *AuthService) issueTokens(
	ctx context.Context,
	userID primitive.ObjectID,
	deviceInfo string,
) (accessToken string, refreshToken string, err error) {
	accessToken, err = utils.GenerateAccessToken(userID.Hex(), s.cfg.JWTSecret, s.cfg.AccessTokenTTL)
	if err != nil {
		return "", "", err
	}

	rawRefresh := utils.GenerateRandomToken()
	refreshHash := utils.HashToken(rawRefresh)

	rt := &models.RefreshToken{
		ID:         primitive.NewObjectID(),
		UserID:     userID,
		TokenHash:  refreshHash,
		DeviceInfo: deviceInfo,
		ExpiresAt:  time.Now().UTC().Add(s.cfg.RefreshTokenTTL),
		Revoked:    false,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.tokenRepo.Create(ctx, rt); err != nil {
		return "", "", err
	}

	return accessToken, rawRefresh, nil
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(s.cfg.JWTSecret), nil
		},
	)

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	return claims.Subject, nil
}

