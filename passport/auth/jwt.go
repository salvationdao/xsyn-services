package auth

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"time"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/terror/v2"
)

func GenerateJWT(tokenID string, u types.User, deviceName string, expireInDays int, jwtKey []byte) (jwt.Token, func(jwt.Token, bool, []byte) ([]byte, error), error) {
	token := openid.New()
	token.Set("user-id", u.ID)
	token.Set(openid.JwtIDKey, tokenID)
	token.Set(openid.ExpirationKey, time.Now().AddDate(0, 0, expireInDays))
	token.Set("device", deviceName)

	if u.FactionID.Valid {
		token.Set("faction-id", u.FactionID.String)
	}

	if u.PublicAddress.Valid {
		token.Set("public-address", u.PublicAddress.String)
	}

	sign := func(t jwt.Token, encryptToken bool, encryptKey []byte) ([]byte, error) {
		if !encryptToken {
			return jwt.Sign(t, jwa.HS256, jwtKey)
		}

		// sign
		signedJWT, err := jwt.Sign(t, jwa.HS256, jwtKey)
		if err != nil {
			passlog.L.Err(err).Str("generate_jwt", tokenID).Msg("Failed to generate token")
			return nil, terror.Error(err, "Failed to sign token")
		}

		// then encrypt
		encryptedAndSignedToken, err := encrypt(encryptKey, signedJWT)
		if err != nil {
			passlog.L.Err(err).Str("generate_jwt", tokenID).Msg("Failed to generate token")
			return nil, terror.Error(err, "Failed to encrypt token")
		}

		return encryptedAndSignedToken, nil
	}
	return token, sign, nil
}

func encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, terror.Error(err, "Failed to encrypt token")
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		passlog.L.Err(err).Str("encrypt", string(key)).Msg("Failed to encrypt token")
		return nil, terror.Error(err, "Failed to encrypt token")
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		passlog.L.Err(err).Str("decrypt", string(key)).Msg("Failed to decrypt token")
		return nil, terror.Error(err, "Failed to decrypt token")
	}
	if len(text) < aes.BlockSize {
		return nil, terror.Error(fmt.Errorf("ciphertext too short"), "Ciphertext too short, please try login in again")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		passlog.L.Err(err).Str("decrypt", string(key)).Msg("Failed to decrypt token")
		return nil, terror.Error(err, "Failed to decrypt token")
	}
	return data, nil
}

// ReadJWT grabs the user from the token
func ReadJWT(tokenB []byte, decryptToken bool, decryptKey, jwtKey []byte) (jwt.Token, error) {
	if !decryptToken {
		token, err := jwt.Parse(tokenB, jwt.WithVerify(jwa.HS256, jwtKey))
		if err != nil {
			return nil, terror.Error(err, "Token verification failed")
		}
		if token.Expiration().Before(time.Now()) {
			return token, terror.Error(err, "Token expired")
		}

		return token, nil
	}

	decrpytedToken, err := decrypt(decryptKey, tokenB)
	if err != nil {
		passlog.L.Err(err).Str("decrypt", string(tokenB)).Msg("Failed to decrypt token")
		return nil, terror.Error(err, "Error decrypting JWT token")
	}

	return ReadJWT(decrpytedToken, false, nil, jwtKey)
}
