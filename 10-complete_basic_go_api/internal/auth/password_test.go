package auth

import "testing"

func TestHashAndVerify(t *testing.T) {
	hash, err := HashPassword("correct-horse")
	if err != nil {
		t.Fatal(err)
	}

	ok, err := VerifyPassword("correct-horse", hash)
	if err != nil || !ok {
		t.Fatalf("verify match: ok=%v err=%v", ok, err)
	}

	ok, err = VerifyPassword("wrong-password", hash)
	if err != nil || ok {
		t.Fatalf("verify mismatch: ok=%v err=%v", ok, err)
	}
}

func TestHashRejectsEmptyAndHuge(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("expected empty password to fail")
	}
	if _, err := HashPassword(string(make([]byte, 257))); err == nil {
		t.Fatal("expected huge password to fail")
	}
}

func TestVerifyRejectsTamperedHash(t *testing.T) {
	hash, err := HashPassword("correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	tampered := hash[:len(hash)-1] + "A"
	if _, err := VerifyPassword("correct-horse", tampered); err == nil {
		ok, verifyErr := VerifyPassword("correct-horse", tampered)
		if verifyErr == nil && ok {
			t.Fatal("tampered hash was accepted")
		}
	}

	if _, err := VerifyPassword("correct-horse", "$argon2id$v=19$m=999999,t=2,p=1$c2FsdA$aGFzaA"); err == nil {
		t.Fatal("expected oversized parameters to fail")
	}
}
