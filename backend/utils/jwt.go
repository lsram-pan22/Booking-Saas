package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleSuperAdmin = "super_admin"
	RoleOwner      = "owner"
	RoleStaff      = "staff"
)

type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

// HashPassword เข้ารหัส Password ก่อนบันทึกลง Database
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CheckPassword ตรวจสอบ Password กับ Hash ที่เก็บใน Database
func CheckPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)

	return err == nil
}

func GenerateJWT(
	secret string,
	expiresIn time.Duration,
	userID int64,
	email string,
	role string,
) (string, error) {

	if secret == "" {
		return "", fmt.Errorf("JWT secret is empty")
	}

	if expiresIn <= 0 {
		return "", fmt.Errorf("JWT expiration must be greater than zero")
	}

	now := time.Now()

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "booking-saas",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("generate JWT: %w", err)
	}

	return signedToken, nil
}
