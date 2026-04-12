package testutil

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/kntjspr/fullstack-golang-next-template/internal/auth"
	"github.com/kntjspr/fullstack-golang-next-template/internal/models"
)

// CreateTestUser inserts a user fixture and returns it.
func CreateTestUser(t *testing.T, db *gorm.DB, overrides map[string]any) models.User {
	t.Helper()

	now := time.Now().UTC().UnixNano()
	user := models.User{
		ID:           fmt.Sprintf("test-user-%d", now),
		Email:        fmt.Sprintf("user-%d@example.com", now),
		Name:         "Test User",
		PasswordHash: "test-password-hash",
		Role:         "user",
	}

	for key, value := range overrides {
		switch key {
		case "id":
			user.ID, _ = value.(string)
		case "email":
			user.Email, _ = value.(string)
		case "name":
			user.Name, _ = value.(string)
		case "password_hash":
			user.PasswordHash, _ = value.(string)
		case "role":
			user.Role, _ = value.(string)
		default:
			t.Fatalf("unsupported override key: %s", key)
		}
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return user
}

// CreateTestAdminUser inserts an admin fixture and returns it.
func CreateTestAdminUser(t *testing.T, db *gorm.DB) models.User {
	t.Helper()
	return CreateTestUser(t, db, map[string]any{"role": "admin"})
}

// GenerateTestToken creates a signed auth token for tests.
func GenerateTestToken(t *testing.T, userID, role string) string {
	t.Helper()

	token, err := auth.GenerateToken(userID, role)
	if err != nil {
		t.Fatalf("generate test token: %v", err)
	}

	return token
}
