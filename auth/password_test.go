package auth

import "testing"

func TestPasswordHash(t *testing.T) {
	password := "Cuppie123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if hash == password {
		t.Fatal("Password should not be stored as plain text")
	}

	if err := CheckPassword(password, hash); err != nil {
		t.Fatal("Correct password should match")
	}

	if err := CheckPassword("WrongPassword", hash); err == nil {
		t.Fatal("Incorrect password should not match")
	}
}
