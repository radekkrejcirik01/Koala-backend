package middleware

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	jtoken "github.com/golang-jwt/jwt/v4"
	"github.com/radekkrejcirik01/Koala-backend/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateJWT creates a token for the user's current token version,
// increasing the version in users table revokes all older tokens
func CreateJWT(username string, tokenVersion int64) (string, error) {
	claims := jtoken.MapClaims{
		"username": username,
		"created":  time.Now(),
		"version":  tokenVersion,
	}

	// Create token
	token := jtoken.NewWithClaims(jtoken.SigningMethodHS512, claims)

	// Generate encoded token and send it as response.
	return token.SignedString([]byte(database.GetJWTSecret()))
}

func Authorize(c *fiber.Ctx) (string, error) {
	// Get the JWT token from the Authorization header
	authHeader := c.Get("Authorization")
	tokenString := ""

	// Check if the Authorization header is present and starts with "Bearer "
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		// Return an error if the Authorization header is missing or in an invalid format
		return "", fiber.ErrUnauthorized
	}

	// Get the key or public key used for validating the signature
	key := []byte(database.GetJWTSecret())

	// Validate the signature and additional claims
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method and return the key
		if token.Method != jwt.SigningMethodHS512 {
			return nil, fiber.ErrUnauthorized
		}
		return key, nil
	})

	if err != nil {
		// Return an error if the JWT signature is invalid or claims are invalid
		return "", fiber.ErrUnauthorized
	}

	// Validate additional claims as per your requirements
	username, ok := claims["username"].(string)
	if !ok {
		return "", fiber.ErrUnauthorized
	}

	// Tokens from before token versions have no version claim, they count as version 0
	version, _ := claims["version"].(float64)

	var tokenVersion int64
	if err := database.DB.
		Table("users").
		Select("token_version").
		Where("username = ?", username).
		Take(&tokenVersion).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fiber.ErrUnauthorized
		}
		return "", err
	}

	if int64(version) != tokenVersion {
		return "", fiber.ErrUnauthorized
	}

	return username, nil
}

// HashPassword hashes the password with bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CheckPassword compares the password with its bcrypt or legacy sha512 hash,
// legacy is true when the hash should be replaced with a bcrypt one
func CheckPassword(hash, password string) (ok bool, legacy bool) {
	if strings.HasPrefix(hash, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil, false
	}

	legacyHash := getLegacyHashPassword(password)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(legacyHash)) == 1, true
}

// getLegacyHashPassword is the unsalted sha512 hash used before bcrypt
func getLegacyHashPassword(password string) string {
	hash := sha512.New()
	hash.Write([]byte(password))
	hashBytes := hash.Sum(nil)

	return hex.EncodeToString(hashBytes)
}
