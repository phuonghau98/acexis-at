package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"
)

type user struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type userWithID struct {
	ID interface{} `json:"_id,omitempty"`
}

type users []user

type loginResponse struct {
	Token string `json:"token"`
}

var testUsers = users{
	user{"name01", "password01", "username01"},
	user{"name02", "passowrd02", "username02"},
}

var registerPath = func() string { return fmt.Sprintf("http://%s:%d/auth/register", *host, *port) }
var userRoute = func() string { return fmt.Sprintf("http://%s:%d/user/", *host, *port) }
var loginPath = func() string { return fmt.Sprintf("http://%s:%d/auth/login", *host, *port) }
var privateResource = func() string { return fmt.Sprintf("http://%s:%d/product/protected", *host, *port) }
var publicResource = func() string { return fmt.Sprintf("http://%s:%d/product/public", *host, *port) }

func checkIfPortIsOpen(host string, port int) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), time.Second*2)

	if err != nil {
		return err
	}
	if conn != nil {
		conn.Close()
	}
	return nil
}

func userRegistration(user user) (*http.Response, error) {
	url := registerPath()
	bodyJSON := []byte(fmt.Sprintf(`{ "name": "%s", "password": "%s", "username": "%s" }`, user.Name, user.Password, user.Username))
	body := bytes.NewBuffer(bodyJSON)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func checkIfICanRegisterAnAccount(reqUser user) error {
	rsp, err := userRegistration(reqUser)
	if err != nil {
		return (err)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 204 {
		return fmt.Errorf("Expected response status code 204, received %d", rsp.StatusCode)
	}
	return nil
}

func checkUserDuplication(dupUser user) error {
	rsp1, err1 := userRegistration(dupUser)
	if err1 != nil {
		return err1
	}
	time.Sleep(time.Second * 3)
	rsp2, err2 := userRegistration(dupUser)
	if err2 != nil {
		return err2
	}

	if rsp2.StatusCode != 409 {
		return fmt.Errorf("Username duplication check, expect status code of 409 on the second response, received %d", rsp2.StatusCode)
	}
	defer rsp1.Body.Close()
	defer rsp2.Body.Close()
	return nil
}

func checkIfPasswordHashed(reqUser user) error {
	reqPath := userRoute() + reqUser.Username
	req, _ := http.NewRequest("GET", reqPath, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	var rspUser user
	err = json.Unmarshal(body, &rspUser)
	if err != nil {
		return fmt.Errorf("No user received from %s, check the return body", reqPath)
	}
	if len(rspUser.Username) == 0 {
		return fmt.Errorf("Cannot get user, check the body: %v", string(body))
	}

	if rspUser.Password == testUsers[0].Password {
		return fmt.Errorf("Password is not hashed")
	}
	return nil
}

func checkIfICanLogin(testUser user) (string, error) {
	url := loginPath()
	bodyJSON := []byte(fmt.Sprintf(`{ "username": "%s", "password": "%s" }`, testUser.Username, testUser.Password))
	body := bytes.NewBuffer(bodyJSON)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	var receivedToken loginResponse
	json.Unmarshal(rspBody, &receivedToken)
	if len(receivedToken.Token) == 0 {
		return "", fmt.Errorf("Cannot rspUserreceived token, check body: %v", string(rspBody))
	}
	return receivedToken.Token, nil
}

func checkPrivateAndPublicRoute(token string, userID string) error {
	// Check if private route worked
	publicExpectedStr := "public content"
	req, _ := http.NewRequest("GET", publicResource(), nil)
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil || rsp.StatusCode == 404 {
		return fmt.Errorf("Cannot process to %s", publicResource())
	}
	rspBody, _ := ioutil.ReadAll(rsp.Body)
	if string(rspBody) != publicExpectedStr {
		return fmt.Errorf("Expected: %s, received: %s", publicExpectedStr, rspBody)
	}
	// Check if private route did not work without token
	req, _ = http.NewRequest("GET", privateResource(), nil)
	rsp, err = client.Do(req)
	if err != nil || rsp.StatusCode == 404 {
		return fmt.Errorf("Cannot process to %s", privateResource())
	}
	if rsp.StatusCode != 403 {
		return fmt.Errorf("Expected status code of 403 without token when processing to %s, received status code of %d", privateResource(), rsp.StatusCode)
	}
	defer rsp.Body.Close()
	// Check if private route did work with token (s3cr3t key)
	expectedStr := "private content of " + userID
	req, _ = http.NewRequest("GET", privateResource(), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rsp, err = client.Do(req)
	if err != nil || rsp.StatusCode == 404 {
		return fmt.Errorf("Cannot process to %s, Error: %s", privateResource(), err)
	}
	rspBody, _ = ioutil.ReadAll(rsp.Body)

	if string(rspBody) != expectedStr {
		return fmt.Errorf("Expected a response body of '%s' when processing to %s, received: '%s'", expectedStr, privateResource(), string(rspBody))
	}
	defer rsp.Body.Close()
	// Check if private route did not work with fake token (s3cr3t key)
	req, _ = http.NewRequest("GET", privateResource(), nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJhVXNlcklkIiwiaWF0IjoxNTE2MjM5MDIyfQ.uGvZ9Z_GASpLxhf_E4aft06kXki-JNDxxl_yKERc4-Y")
	rsp, err = client.Do(req)
	if err != nil || rsp.StatusCode == 404 {
		return fmt.Errorf("Cannot process to %s, Error: %s", privateResource(), err)
	}
	if rsp.StatusCode != 403 {
		return fmt.Errorf("Expected status code of 403 when processing to %s using a fake token, but received status code of %d", privateResource, rsp.StatusCode)
	}
	defer rsp.Body.Close()
	return nil
}
