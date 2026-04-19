package auth

import "testing"

func TestHashAndComparePassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "correct-password" {
		t.Fatal("password hash should not equal plaintext")
	}

	if err := ComparePassword(hash, "correct-password"); err != nil {
		t.Fatalf("compare password: %v", err)
	}
	if err := ComparePassword(hash, "wrong-password"); err == nil {
		t.Fatal("expected compare failure for wrong password")
	}
}

func TestHashPassword_EmptyInput(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected error for empty password")
	}
}
