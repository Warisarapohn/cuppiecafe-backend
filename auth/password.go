package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword converts a plain password into a secure hash.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// CheckPassword compares a password with its stored hash.
func CheckPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
}
