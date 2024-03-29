package crypto_test

import (
	"testing"
	"xsyn-services/passport/crypto"
)

func TestCrypto(t *testing.T) {
	t.Run("compare_password", func(t *testing.T) {
		password := "NinjaDojo_!"
		hashed := crypto.HashPassword(password)
		err := crypto.ComparePassword(hashed, password)
		if err != nil {
			t.Fatal(err)
		}
	})
}
