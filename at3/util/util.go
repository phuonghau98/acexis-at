package util

import (
	"athelper/submit"
	"fmt"
	"os"

	"github.com/dgrijalva/jwt-go"
)

func SignWithID(ID string) (string, error) {
	secret := []byte("s3cr3t")
	type claim struct {
		AuthorID string `json:"authorID"`
		jwt.StandardClaims
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim{AuthorID: ID})
	r, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return r, nil
}

func Submit(r bool) {
	trainingServer := os.Getenv("TRAINING_SERVER")
	err := submit.SubmitReport(trainingServer, "3", r)
	if err != nil {
		fmt.Println("Cannot submit to training server, contact admin\n", err.Error())
	}
}
