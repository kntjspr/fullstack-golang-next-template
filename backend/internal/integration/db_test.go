//go:build integration

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/create-go-app/chi-go-template/internal/models"
	"github.com/create-go-app/chi-go-template/internal/testutil"
)

func TestMigrateUpDown(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	if !db.Migrator().HasTable(&models.User{}) {
		t.Fatal("users table should exist after AutoMigrate")
	}

	if err := db.Migrator().DropTable(&models.User{}); err != nil {
		t.Fatalf("drop users table: %v", err)
	}

	if db.Migrator().HasTable(&models.User{}) {
		t.Fatal("users table should not exist after drop")
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("remigrate users table: %v", err)
	}

	if !db.Migrator().HasTable(&models.User{}) {
		t.Fatal("users table should exist after remigrate")
	}
}

func TestMigrateIdempotent(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("first automigrate: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("second automigrate: %v", err)
	}
}

func TestUserCRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)
	defer testutil.TeardownTestDB(t, db)

	userID := fmt.Sprintf("00000000-0000-0000-0000-%012d", time.Now().UTC().UnixNano()%1_000_000_000_000)
	email := fmt.Sprintf("integration-%d@example.com", time.Now().UTC().UnixNano())

	user := models.User{
		ID:           userID,
		Email:        email,
		Name:         "Integration User",
		PasswordHash: "hashed-password",
		Role:         "user",
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	var fetched models.User
	if err := db.First(&fetched, "id = ?", userID).Error; err != nil {
		t.Fatalf("fetch created user: %v", err)
	}

	if fetched.Email != email {
		t.Fatalf("unexpected email: got %q, want %q", fetched.Email, email)
	}

	if err := db.Model(&fetched).Update("role", "admin").Error; err != nil {
		t.Fatalf("update user role: %v", err)
	}

	var updated models.User
	if err := db.First(&updated, "id = ?", userID).Error; err != nil {
		t.Fatalf("fetch updated user: %v", err)
	}

	if updated.Role != "admin" {
		t.Fatalf("unexpected role: got %q, want %q", updated.Role, "admin")
	}

	if err := db.Delete(&models.User{}, "id = ?", userID).Error; err != nil {
		t.Fatalf("delete user: %v", err)
	}

	var remaining int64
	if err := db.Model(&models.User{}).Where("id = ?", userID).Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining users: %v", err)
	}

	if remaining != 0 {
		t.Fatalf("expected user to be deleted, got %d matching rows", remaining)
	}
}
