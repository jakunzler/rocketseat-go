package validator

import (
	"testing"
)

func TestCheckFieldKeepsFirstError(t *testing.T) {
	var eval Evaluator
	eval.CheckField(false, "email", "first")
	eval.CheckField(false, "email", "second")
	if eval["email"] != "first" {
		t.Fatalf("email error = %q", eval["email"])
	}
}

func TestPasswordOK(t *testing.T) {
	tests := []struct {
		password string
		ok       bool
	}{
		{password: "short-pass", ok: false},
		{password: "correct-horse", ok: true},
		{password: "senha-com-acentos-ção", ok: true},
		{password: string(make([]byte, 257)), ok: false},
	}
	for _, tt := range tests {
		if got := PasswordOK(tt.password); got != tt.ok {
			t.Fatalf("PasswordOK(%q) = %v, want %v", tt.password, got, tt.ok)
		}
	}
}

func TestMatchesEmail(t *testing.T) {
	if !Matches("ana@example.com", EmailRX) {
		t.Fatal("expected valid email")
	}
	if Matches("not-an-email", EmailRX) {
		t.Fatal("expected invalid email")
	}
}

func TestNotBlankAndBounds(t *testing.T) {
	if NotBlank("  ") {
		t.Fatal("blank string should fail")
	}
	if !MinChars("abc", 3) || !MaxChars("abc", 3) {
		t.Fatal("bounds failed")
	}
	if MaxChars("áé", 1) {
		t.Fatal("rune count should reject two characters")
	}
}
