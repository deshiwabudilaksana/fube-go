package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/deshiwabudilaksana/fube-go/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	db *mongo.Database
}

func NewService(db *mongo.Database) *Service {
	return &Service{db: db}
}

type RegisterParams struct {
	Username    string
	Email       string
	Password    string
	CompanyName string
}

func (s *Service) RegisterUser(ctx context.Context, params RegisterParams) (*models.UserDoc, error) {
	usersColl := s.db.Collection("users")
	vendorsColl := s.db.Collection("vendors")

	// Check if user already exists
	var existing models.UserDoc
	err := usersColl.FindOne(ctx, bson.M{"email": params.Email}).Decode(&existing)
	if err == nil {
		return nil, ErrUserExists
	}
	if err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := primitive.NewObjectID()
	vendorID := primitive.NewObjectID()

	// Create user
	user := models.UserDoc{
		ID:        userID,
		Username:  params.Username,
		Email:     params.Email,
		Password:  string(hashedPassword),
		Role:      models.RoleOwner,
		VendorID:  vendorID.Hex(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = usersColl.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	// Create default vendor doc
	vendor := models.VendorDoc{
		ID:          vendorID,
		CompanyName: params.CompanyName,
		Email:       params.Email,
		UserID:      userID.Hex(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err = vendorsColl.InsertOne(ctx, vendor)
	if err != nil {
		return nil, fmt.Errorf("failed to insert vendor: %w", err)
	}

	return &user, nil
}

func (s *Service) LoginUser(ctx context.Context, email, password string) (string, error) {
	usersColl := s.db.Collection("users")

	var user models.UserDoc
	err := usersColl.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Generate JWT
	token, err := s.GenerateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return token, nil
}

func (s *Service) GenerateJWT(user models.UserDoc) (string, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET_KEY not set")
	}

	claims := jwt.MapClaims{
		"user_id":   user.ID.Hex(),
		"vendor_id": user.VendorID,
		"username":  user.Username,
		"email":     user.Email,
		"role":      string(user.Role),
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
