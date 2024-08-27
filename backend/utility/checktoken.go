package utility

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
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
	tokenJson := gjson.New(claims["https://api.openai.com/profile"])
	// tokenJson.Dump()
	email := tokenJson.Get("email").String()
	if email == "" {
		return gerror.New("Invalid token")
	}
	// 验证是否过期
	// err = claims.Valid()
	// if err != nil {
	// 	return err
	// }

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
	// g.Dump(claims)
	if email == "" {
		return email, gerror.New("Invalid token")
	}

	// // 验证是否过期
	// if !claims.Valid() {
	// 	return email, errors.New("token is expired")
	// }
	return email, nil
}
