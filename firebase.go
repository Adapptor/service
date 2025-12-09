package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type GoogleCredentials struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
}

func ReadGoogleAuth(path string) (GoogleCredentials, error) {
	var (
		file []byte
		err  error
		gc   GoogleCredentials
	)

	if file, err = ioutil.ReadFile(path); err == nil {
		err = json.Unmarshal(file, &gc)
	}

	return gc, err
}

func ParseFirebaseToken(tokenString, secret string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}
