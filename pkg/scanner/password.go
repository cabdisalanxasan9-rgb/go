package scanner

import (
	"math"
	"unicode"
)

func CheckPasswordStrength(password string) *PasswordStrengthResult {
	length := len(password)
	poolSize := 0

	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	if hasLower {
		poolSize += 26
	}
	if hasUpper {
		poolSize += 26
	}
	if hasDigit {
		poolSize += 10
	}
	if hasSymbol {
		poolSize += 33
	}

	entropy := 0.0
	if poolSize > 0 && length > 0 {
		entropy = float64(length) * math.Log2(float64(poolSize))
	}

	strength := "weak"
	switch {
	case entropy >= 80:
		strength = "very-strong"
	case entropy >= 60:
		strength = "strong"
	case entropy >= 40:
		strength = "moderate"
	}

	warnings := make([]string, 0)
	if length < 12 {
		warnings = append(warnings, "password should be at least 12 characters")
	}
	if !hasLower || !hasUpper || !hasDigit || !hasSymbol {
		warnings = append(warnings, "use upper/lowercase letters, numbers, and symbols")
	}

	return &PasswordStrengthResult{
		Length:      length,
		EntropyBits: entropy,
		Strength:    strength,
		Warnings:    warnings,
	}
}
