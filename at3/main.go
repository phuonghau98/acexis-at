package main

import (
	"at3/query"
	"at3/subscription"
	"at3/util"
	"athelper/logger"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/machinebox/graphql"
)

var wg sync.WaitGroup

func main() {

	appHost := os.Getenv("APP_HOST")
	appPort := os.Getenv("APP_PORT")
	// CONFIG DECLARATIONS
	graphqlWsNP := fmt.Sprintf("ws://%s:%s/graphql", appHost, appPort)
	graphqlHTTPNP := fmt.Sprintf("http://%s:%s/graphql", appHost, appPort)
	userBConn, err := subscription.WsConnect(graphqlWsNP)
	userCConn, _ := subscription.WsConnect(graphqlWsNP)
	userDConn, _ := subscription.WsConnect(graphqlWsNP)
	userEConn, _ := subscription.WsConnect(graphqlWsNP)
	// backboneCtx := context.Background()
	// END OF CONFIG DECLARATIONS
	logger.NewCase("Establish websocket connection")
	logger.NewInfo(fmt.Sprintf("Establishing subscription connection to %s", graphqlWsNP))
	if err != nil {
		util.Submit(false)
		logger.NewError(fmt.Sprintf("Cannot establish websocket connection to %s, detail: \n %v", graphqlWsNP, err))
	}
	logger.NewSuccess("Connection established")
	defer userBConn.Close()
	defer userCConn.Close()
	defer userDConn.Close()
	defer userEConn.Close()
	sessionB := &subscription.Session{
		Ws: userBConn,
	}
	sessionC := &subscription.Session{
		Ws: userCConn,
	}
	userDConnErr := make(chan error)
	userEConnErr := make(chan error)
	sessionD := &subscription.Session{
		Ws:      userDConn,
		ErrChan: userDConnErr,
	}
	sessionE := &subscription.Session{
		Ws:      userEConn,
		ErrChan: userEConnErr,
	}
	logger.NewCase("Users can send and receive messages correctly")
	// Init graphql client
	client := graphql.NewClient(graphqlHTTPNP)
	// End of init graphql client
	// TOKEN DECLARATION
	// NOTE: 3113
	userAToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyYSIsInByaXZpbGVnZXMiOlsiMzExMyJdfQ.ZE2nH80O6HQ33zKmOOk7IdS3y-V2IHpzg_wQDZ0rIZE"
	// NOTE: 3113
	userBToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyYiIsInByaXZpbGVnZXMiOlsiMzExMyJdfQ.mzePr1FEw4OGIgsop0FiBmQR3HWcfnOjiBn4jglWjXE"
	// NOTE: 3113, 3114
	userCToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyYyIsInByaXZpbGVnZXMiOlsiMzExMyIsIjMxMTQiXX0.36ZGjDGcHBDLMoglG1Vz7Ci8cWFYpeYlqbYSx98QFkc"
	// NOTE: 3115
	userEToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySUQiOiJ1c2VyZSIsInByaXZpbGVnZXMiOlsiMzExNSJdfQ.Fwws_3m6q5VTMGEDwUqMM9tjqa3s9NgHrr9WNPOqDtE"
	// END OF TOKEN DECLARATION
	roomAID := "3113"
	roomBID := "3114"
	logger.NewInfo(fmt.Sprintf("User-A creating message with authorization header %s with roomID %s", newAuthHeader(userAToken), roomAID))
	messageMutationReq := query.NewMessageMutationRequest()
	var message1Response *query.MessageResponse
	messageMutationReq.Var("content", "Hello message from a")
	messageMutationReq.Var("roomID", roomAID)
	messageMutationReq.Header.Add("Authorization", newAuthHeader(userAToken))
	if err := client.Run(context.TODO(), messageMutationReq, &message1Response); err != nil {
		util.Submit(false)
		logger.NewError(fmt.Sprintf("User-A cannot create message with authorization header: %s", newAuthHeader(userAToken)))
	}
	logger.NewSuccess(fmt.Sprintf("User-A created a message \n response: %v", message1Response))
	logger.NewInfo(fmt.Sprintf("User-B starts subscribing to all messages created with roomID: %s", roomAID))

	wg.Add(4)
	msgOfASentContent := "Hello message from a an expect b can receive"
	go func() {
		sessionB.Ws.WriteJSON(&subscription.OperationMessage{Type: subscription.ConnectionInitMsg, Payload: json.RawMessage(fmt.Sprintf("{\"Authorization\": \"%s\"}", newAuthHeader(userBToken)))})
		sessionB.ReadOp()
		// ctx, _ := context.WithTimeout(backboneCtx, time.Second*3)
		queryStr := fmt.Sprintf(`{"query": "subscription { messageCreated (roomID: \"%s\") { _id content createdAt roomID createdBy } }"}`, roomAID)
		subscriptionFuture, _ := sessionB.Subscribe(queryStr)
		select {
		case s := <-subscriptionFuture:
			var res query.MessageSubscriptionResponse
			parseErr := json.Unmarshal([]byte(s), &res)
			if parseErr != nil {
				fmt.Printf("Cannot parse the response, check response: %s", err.Error())
			}
			time.Sleep(time.Second * 1)
			if res.Data.MessageCreated.Content == msgOfASentContent {
				logger.NewSuccess("User-B received a message from user-A")
			} else {
				util.Submit(false)
				logger.NewError(fmt.Sprintf("User-B expect to receive a message with content: %s, but received: %s", msgOfASentContent, res.Data.MessageCreated.Content))
			}
			break
		case <-time.After(time.Second * 5):
			util.Submit(false)
			logger.NewError(fmt.Sprintf("User-B has been waited for 5 seconds but no message received"))
		}
		defer wg.Done()
	}()
	// USER C Subscribe to another room
	logger.NewInfo(fmt.Sprintf("User-C starts subscribing to all messages created with roomID: %s", roomBID))
	go func() {
		sessionC.Ws.WriteJSON(&subscription.OperationMessage{Type: subscription.ConnectionInitMsg, Payload: json.RawMessage(fmt.Sprintf("{\"Authorization\": \"%s\"}", newAuthHeader(userCToken)))})
		sessionC.ReadOp()
		// ctx, _ := context.WithTimeout(backboneCtx, time.Second*3)
		queryStr := fmt.Sprintf(`{"query": "subscription { messageCreated (roomID: \"%s\") { _id content createdAt roomID createdBy } }"}`, roomBID)
		subscriptionFuture, _ := sessionC.Subscribe(queryStr)
		select {
		case <-subscriptionFuture:
			util.Submit(false)
			logger.NewError(fmt.Sprintf("User-C subscribed to messages with roomID: %s, but received a message with roomID: %s", roomBID, roomAID))
			break
		case <-time.After(time.Second * 3):
			logger.NewSuccess(fmt.Sprintf("User-C did not receive any message sent from user-A to roomID: %s", roomAID))
		}
		defer wg.Done()
	}()
	// SLEEP 1 second
	// User-A sending message
	time.Sleep(time.Second * 1)
	logger.NewInfo(fmt.Sprintf("User-A is sending a message with roomID: %s", roomAID))
	msgMutationReq := query.NewMessageMutationRequest()
	msgMutationReq.Header.Add("Authorization", newAuthHeader(userAToken))
	messageMutationReq.Var("content", msgOfASentContent)
	messageMutationReq.Var("roomID", roomAID)
	if err := client.Run(context.TODO(), messageMutationReq, nil); err != nil {
		util.Submit(false)
		logger.NewError("User-A cannot send message")
	} else {
		logger.NewSuccess("User-A's message sent")
	}
	// User-D cannot subscribe to server without authorization header
	logger.NewInfo(fmt.Sprintf("User-D starts subscribing to all messages created with roomID: %s without Authorization header (not authenticated)", roomAID))
	go func() {
		sessionD.Ws.WriteJSON(&subscription.OperationMessage{Type: subscription.ConnectionInitMsg, Payload: json.RawMessage("{}")})
		sessionD.ReadOp()
		queryStr := fmt.Sprintf(`{"query": "subscription { messageCreated (roomID: \"%s\") { _id content createdAt roomID createdBy } }"}`, roomBID)
		sessionD.Subscribe(queryStr)
		time.Sleep(time.Second * 1)
		select {
		case <-userDConnErr:
			logger.NewSuccess("User-D cannot subscribe to messages without Authorization header")
			break
		case <-time.After(time.Second * 3):
			util.Submit(false)
			logger.NewError("User-D subscrition did not fail after 3 seconds")
		}
		defer wg.Done()
	}()
	// User-E Case
	go func() {
		sessionE.Ws.WriteJSON(&subscription.OperationMessage{Type: subscription.ConnectionInitMsg, Payload: json.RawMessage(fmt.Sprintf("{\"Authorization\": \"%s\"}", newAuthHeader(userEToken)))})
		sessionE.ReadOp()
		queryStr := fmt.Sprintf(`{"query": "subscription { messageCreated (roomID: \"%s\") { _id content createdAt roomID createdBy } }"}`, roomAID)
		subscriptionFuture, _ := sessionE.Subscribe(queryStr)
		select {
		case s := <-subscriptionFuture:
			fmt.Println(s)
			util.Submit(false)
			logger.NewError("User-E still receive the message from A")
			break
		case <-time.After(time.Second * 5):
			logger.NewSuccess(fmt.Sprintf("User-E cannot subscribe to messages with roomID %s due to his privileges", roomAID))
		}
		defer wg.Done()
	}()
	time.Sleep(1 * time.Second)
	logger.NewInfo(fmt.Sprintf("User-A is sending a message with roomID: %s", roomAID))
	messageMutationReq.Var("content", msgOfASentContent)
	messageMutationReq.Var("roomID", roomAID)
	if err := client.Run(context.TODO(), messageMutationReq, nil); err != nil {
		util.Submit(false)
		logger.NewError("User-A cannot send message")
	} else {
		logger.NewSuccess("User-A's message sent")
	}
	wg.Wait()
	util.Submit(true)
}

func newAuthHeader(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
