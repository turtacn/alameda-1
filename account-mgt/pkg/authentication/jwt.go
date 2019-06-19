// Package authentication : original author Nick Vellios, http://github.com/nickvellios/gojwt/
package authentication

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	//"errors"
	"strings"
	"time"
	//Logger "github.com/containers-ai/alameda/pkg/utils/log"
	Errors "github.com/containers-ai/alameda/pkg/utils/errors"
)

// JwtSalt is used when signing a JWT
var JwtTimeFormat = "2006-01-02T15:04:05Z"
var JwtSalt = "Fed.ai"
var defExpDuration = 3600
var defRenewDuration = 60

// JwtGenerate : Generate compiles and signs a JWT from a claim and an expiration time in seconds from current time.
func JwtGenerate(claim map[string]string, exp int) string {
	if exp <= 0 {
		exp = defExpDuration
	}
	ex := time.Now().Add(time.Second * time.Duration(exp))
	expiration := ex.Format(JwtTimeFormat)
	// Build the jwt header by hand since alg and typ aren't going to change (for now)
	header := base64.StdEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT","exp":"` + expiration + `"}`))
	// Build json payload and base64 encode it
	pl2, err := json.Marshal(claim)
	if err != nil {
		scope.Error(err.Error())
		return ""
	}
	payload := base64.StdEncoding.EncodeToString([]byte(pl2))
	// Create a new secret from our JwtSalt and the paylod json string.
	secret := sha256wrapper(JwtSalt + string(pl2))
	// Build signature with the new secret and base64 encode it.
	hash := hmac256(header+"."+payload, secret)
	signature := base64.StdEncoding.EncodeToString([]byte(hash))
	jwt := header + "." + payload + "." + signature
	return jwt
}

// JwtDecode : Decode decodes a JWT and returns the payload as a map[string]string.
func JwtDecode(jwt string) (map[string]string, string, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	header, _ := base64.StdEncoding.DecodeString(parts[0])
	payload, _ := base64.StdEncoding.DecodeString(parts[1])
	signature, _ := base64.StdEncoding.DecodeString(parts[2])
	// JSON decode payload
	var pldat map[string]string
	if err := json.Unmarshal(payload, &pldat); err != nil {
		scope.Error(err.Error())
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	// JSON decode header
	var headdat map[string]interface{}
	if err := json.Unmarshal(header, &headdat); err != nil {
		scope.Error(err.Error())
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	// Extract and parse expiration date from header
	layout := JwtTimeFormat
	exp := headdat["exp"].(string)
	expParsed, err := time.ParseInLocation(layout, exp, time.Now().Location())
	if err != nil {
		scope.Error(err.Error())
		return nil, jwt, Errors.NewError(Errors.ReasonFailedToGenJWT)
	}
	// Check how old the JWT is.  Return an error if it is expired
	now := time.Now()
	if now.After(expParsed) {
		return nil, jwt, Errors.NewError(Errors.ReasonExpiredJWT)
	}
	pldat["expired"] = exp

	// This probably should be one of the first checks, preceding the date check.  If the signature of the JWT doesn't match there is likely fuckery afoot
	ha := hmac256(string(parts[0])+"."+string(parts[1]), sha256wrapper(JwtSalt+string(payload)))
	if ha != string(signature) {
		scope.Error("invalid JWT signature")
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	// JWT is going to expire in 1 minute, return a new JWT
	inOneMin := expParsed.Add((time.Duration)(0-defRenewDuration) * time.Second)
	if now.After(inOneMin) {
		newJwt := JwtGenerate(pldat, 0)
		return pldat, newJwt, nil
	}

	return pldat, jwt, nil
}

// JwtDecode : Decode decodes a JWT and returns the payload as a map[string]string.
func JwtDecodeWithoutCheckExpired(jwt string) (map[string]string, string, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		scope.Error("invalid JWT structure")
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	header, _ := base64.StdEncoding.DecodeString(parts[0])
	payload, _ := base64.StdEncoding.DecodeString(parts[1])
	signature, _ := base64.StdEncoding.DecodeString(parts[2])
	// JSON decode payload
	var pldat map[string]string
	if err := json.Unmarshal(payload, &pldat); err != nil {
		scope.Error(err.Error())
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	// JSON decode header
	var headdat map[string]interface{}
	if err := json.Unmarshal(header, &headdat); err != nil {
		scope.Error(err.Error())
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}

	// This probably should be one of the first checks, preceding the date check.  If the signature of the JWT doesn't match there is likely fuckery afoot
	ha := hmac256(string(parts[0])+"."+string(parts[1]), sha256wrapper(JwtSalt+string(payload)))
	if ha != string(signature) {
		scope.Error("invalid JWT signature")
		return nil, jwt, Errors.NewError(Errors.ReasonInvalidJWT)
	}

	return pldat, jwt, nil
}

// JwtInfo : Decode a JWT and returns the payload as a map[string]string.
func JwtInfo(jwt string) (map[string]string, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		scope.Error("invalid JWT structure")
		return nil, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	header, _ := base64.StdEncoding.DecodeString(parts[0])
	payload, _ := base64.StdEncoding.DecodeString(parts[1])
	signature, _ := base64.StdEncoding.DecodeString(parts[2])
	// JSON decode payload
	var pldat map[string]string
	if err := json.Unmarshal(payload, &pldat); err != nil {
		scope.Error(err.Error())
		return nil, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	// JSON decode header
	var headdat map[string]interface{}
	if err := json.Unmarshal(header, &headdat); err != nil {
		scope.Error(err.Error())
		return nil, Errors.NewError(Errors.ReasonInvalidJWT)
	}
	// Extract and parse expiration date from header
	//layout := timeFormat
	exp := headdat["exp"].(string)
	pldat["expired"] = exp

	// This probably should be one of the first checks, preceding the date check.  If the signature of the JWT doesn't match there is likely fuckery afoot
	ha := hmac256(string(parts[0])+"."+string(parts[1]), sha256wrapper(JwtSalt+string(payload)))
	if ha != string(signature) {
		scope.Error("invalid JWT signature")
		return nil, Errors.NewError(Errors.ReasonInvalidJWT)
	}

	return pldat, nil
}

func hmac256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func sha256wrapper(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
