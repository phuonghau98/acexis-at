package submit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func parseEnv() (map[string]string, error) {
	file, err := os.Open("app/.env")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	parsedMap := map[string]string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), "=")
		if len(s) == 2 {
			parsedMap[strings.ToLower(s[0])] = s[1]
		}
	}
	return parsedMap, nil
}

func SubmitReport(submitServer string, test string, status bool) error {
	bodyMap, err := parseEnv()
	if err != nil {
		fmt.Println(err.Error())
	}
	if status {
		bodyMap["_status"] = strconv.Itoa(1)
	} else {
		bodyMap["_status"] = strconv.Itoa(0)
	}
	if test != "" {
		bodyMap["test"] = test
	} else {
		return fmt.Errorf("No test specified")
	}
	jsonStr, err := json.Marshal(bodyMap)
	bodyBuffer := bytes.NewBuffer(jsonStr)
	req, err := http.NewRequest("POST", submitServer+"/test/submit", bodyBuffer)
	if err != nil {
		return err
	}
	client := &http.Client{}
	rsp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()
	parsedRsp, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	if rsp.StatusCode == 201 {
		fmt.Println("Submit test successfully")
	} else {
		return fmt.Errorf("Submit failed, response: %v", string(parsedRsp))
	}
	return err
}
