package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "time"

	jwt "github.com/golang-jwt/jwt"
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
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

// func GenerateFirebaseToken(serviceAccountEmail, privateKey string, claims jwt.MapClaims) {
//
// 	now := time.Now()
//
// 	expiry := now.Unix() + 60*60
// 	firebaseUid := "dacff6cc-507b-4ab8-8d3c-5f90b8f5b34e"
//
// 	payload := jwt.MapClaims{
// 		"nbf": now.Unix(),
// 		"iss": serviceAccountEmail,
// 		"sub": serviceAccountEmail,
// 		"aud": "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit",
// 		"iat": now.Unix(),
// 		// XXX EXPIRY
// 		"exp": expiry,
// 		"uid": firebaseUid,
// 	}
//
// 	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
//
// 	// // Sign and get the complete encoded token as a string using the secret
// 	// tokenString, err := token.SignedString(hmacSampleSecret)
//
// 	// fmt.Println(tokenString, err)
//
// 	// $payload = array(
// 	//   "iss" => $service_account_email,
// 	//   "sub" => $service_account_email,
// 	//   "aud" => "https://identitytoolkit.googleapis.com/google.identity.identitytoolkit.v1.IdentityToolkit",
// 	//   "iat" => $now_seconds,
// 	//   "exp" => $now_seconds+(60*60),  // Maximum expiration time is one hour
// 	//   "uid" => $uid,
// 	//   "claims" => array(
// 	//     "premium_account" => $is_premium_account
// 	//   )
// 	// return JWT::encode($payload, $private_key, "RS256");
//
// }
