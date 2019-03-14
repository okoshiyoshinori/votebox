package votetoken

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"

	"github.com/okoshiyoshinori/votebox/config"

	"github.com/gin-gonic/gin"

	"github.com/okoshiyoshinori/votebox/logger"

	jwt "github.com/dgrijalva/jwt-go"
	request "github.com/dgrijalva/jwt-go/request"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

func CreateToken(id string) string {

	signBytes, err := ioutil.ReadFile(config.APConfig.KeyPath + "voteApi")
	if err != nil {
		panic(err)
	}
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}
	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		logger.Error.Fatal(err)
	}
	return tokenString
}

func TokenAuthentication(c *gin.Context) (string, error) {

	verifyBytes, err := ioutil.ReadFile(config.APConfig.KeyPath + "voteApi.pkcs8")
	if err != nil {
		panic(err)
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}
	token, err := request.ParseFromRequest(c.Request, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		_, err := token.Method.(*jwt.SigningMethodRSA)
		if !err {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return verifyKey, nil
		}
	})

	if err == nil && token.Valid {
		claims := token.Claims.(jwt.MapClaims)
		return claims["id"].(string), nil
	} else {
		return "", err
	}

}

func CheckToken(c *gin.Context) bool {

	verifyBytes, err := ioutil.ReadFile(config.APConfig.KeyPath + "voteApi.pkcs8")
	if err != nil {
		panic(err)
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}
	token, err := request.ParseFromRequest(c.Request, request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
		_, err := token.Method.(*jwt.SigningMethodRSA)
		if !err {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		} else {
			return verifyKey, nil
		}
	})
	return token.Valid
}
