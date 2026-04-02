package password

import (
	"math"
	"unicode"
)

type Result struct {
	Length      int      `json:"length"`
	EntropyBits float64  `json:"entropy_bits"`
	Strength    string   `json:"strength"`
	Warnings    []string `json:"warnings,omitempty"`
}

func Check(pass string) *Result {
	length := len(pass)
	pool := 0
	var lower, upper, digit, symbol bool

	for _, ch := range pass {
		switch {
		case unicode.IsLower(ch):
			lower = true
		case unicode.IsUpper(ch):
			upper = true
		case unicode.IsDigit(ch):
			digit = true
		default:
			symbol = true
		}
	}

	if lower {
		pool += 26
	}
	if upper {
		pool += 26
	}
	if digit {
		pool += 10
	}
	if symbol {
		pool += 33
	}

	entropy := 0.0
	if length > 0 && pool > 0 {
		entropy = float64(length) * math.Log2(float64(pool))
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
	if !lower || !upper || !digit || !symbol {
		warnings = append(warnings, "use uppercase, lowercase, number, and symbol")
	}

	return &Result{Length: length, EntropyBits: entropy, Strength: strength, Warnings: warnings}
}
