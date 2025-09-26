package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) string {
	pass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(pass)
}


func ComparePassword(password, hPassword string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hPassword), []byte(password)); err != nil {
		return false
	}
	return true
}
