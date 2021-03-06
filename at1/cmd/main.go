package main

import (
	"athelper/logger/logger"
	"athelper/logger/submit"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var port = flag.Int("port", 3000, "Listening port")
var host = flag.String("host", "127.0.0.1", "Hostname")
var trainingServer = flag.String("server", "http://training.phuonghau.com", "Server for submission")

func init() {
	flag.Parse()
}

func submitFail(trainingServer string) {
	err := submit.SubmitReport(trainingServer, "1", false)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	// Check app running
	logger.NewCase("Application operation")
	err := checkIfPortIsOpen(*host, *port)
	if err != nil {
		submitFail(*trainingServer)
		logger.NewError(err.Error())
	} else {
		logger.NewSuccess(fmt.Sprintf("App is running on port %d", *port))
	}
	// User registration
	logger.NewCase("User registration")
	err = checkIfICanRegisterAnAccount(user{"user0001", "user0001", "user0001" + strconv.Itoa(int(time.Now().Unix()))})
	if err != nil {
		submitFail(*trainingServer)
		logger.NewError(err.Error())
	} else {
		logger.NewSuccess(fmt.Sprintf("user registration is available at %s", registerPath()))
		logger.NewSuccess("Modified status: Received status code of 204 instead of 201")
	}
	// // Duplication check
	err = checkUserDuplication(user{"Phuong", "user" + string(time.Now().Unix()), ""})
	if err != nil {
		submitFail(*trainingServer)
		logger.NewError(err.Error())
	} else {
		logger.NewSuccess("Username duplication check")
	}
	// Check if password is hashed
	randUsername := "user0001" + strconv.Itoa(int(time.Now().Unix()))
	pwdHashUser := user{"user0001", "user0001", randUsername}
	checkIfICanRegisterAnAccount(pwdHashUser)
	err = checkIfPasswordHashed(pwdHashUser)
	if err != nil {
		submitFail(*trainingServer)
		logger.NewError(err.Error())
	} else {
		logger.NewSuccess("Password is hashed")
	}
	// User Authentication
	logger.NewCase("User Login")
	tmpUser := user{"Phuong", "hau.phuong", "phuong.hau"}
	checkIfICanRegisterAnAccount(tmpUser)
	token, err := checkIfICanLogin(user{"Phuong", "hau.phuong", "phuong.hau"})
	if err != nil {
		submitFail(*trainingServer)
		logger.NewError(err.Error())
	} else {
		logger.NewSuccess("Received token")
	}

	reqPath := userRoute() + "phuong.hau"
	req, _ := http.NewRequest("GET", reqPath, nil)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	rsp, err := client.Do(req)
	// if err != nil {
	// 	return err
	// }
	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	var rspUser userWithID
	err = json.Unmarshal(body, &rspUser)
	// if len(rspUser.Username) == 0 {
	// 	return fmt.Errorf("Cannot get user, check the body: %v", string(body))
	// }
	err = checkPrivateAndPublicRoute(token, fmt.Sprintf("%v", rspUser.ID))
	if err != nil {
		submitFail(*trainingServer)
		logger.NewError(err.Error())
	} else {
		logger.NewSuccess("Private route and publish route work correctly")
	}
	err = submit.SubmitReport(*trainingServer, "1", true)
	if err != nil {
		fmt.Println(err.Error())
	}
}
