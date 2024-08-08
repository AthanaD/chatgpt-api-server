package utility

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gogf/gf/v2/encoding/gjson"
)

func parseJWTWithoutValidation(tokenString string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	} else {
		return nil, err
	}
}

func CheckAccessToken(tokenString string) error {
	claims, err := parseJWTWithoutValidation(tokenString)
	if err != nil {

		return err
	}
	// 验证是否过期
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return errors.New("token is expired")
	}
	return nil
}

func ParserAccessToken(tokenString string) (email string, err error) {
	claims, err := parseJWTWithoutValidation(tokenString)
	if err != nil {

		return "", err
	}
	tokenJson := gjson.New(claims["https://api.openai.com/profile"])
	// tokenJson.Dump()
	email = tokenJson.Get("email").String()
	// 验证是否过期
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return email, errors.New("token is expired")
	}
	return email, nil
}
