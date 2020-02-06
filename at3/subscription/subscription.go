package subscription

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type Session struct {
	Ws      *websocket.Conn
	ErrChan chan error
}

const (
	ConnectionInitMsg      = "connection_init"      // Client -> Server
	connectionTerminateMsg = "connection_terminate" // Client -> Server
	startMsg               = "start"                // Client -> Server
	stopMsg                = "stop"                 // Client -> Server
	connectionAckMsg       = "connection_ack"       // Server -> Client
	connectionErrorMsg     = "connection_error"     // Server -> Client
	dataMsg                = "data"                 // Server -> Client
	errorMsg               = "error"                // Server -> Client
	completeMsg            = "complete"             // Server -> Client
	//connectionKeepAliveMsg = "ka"                 // Server -> Client  TODO: keepalives
)

type OperationMessage struct {
	Payload json.RawMessage `json:"payload,omitempty"`
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
}

// type PostSubscriptionResponse struct {
// 	Data struct {
// 		Post struct {
// 			Node struct {
// 				ID    string `json:"id"`
// 				Title string `json:"title"`
// 			} `json:"node"`
// 		} `json:"post"`
// 	} `json:"data"`
// }

func WsConnect(url string) (*websocket.Conn, error) {
	headers := make(http.Header)
	headers.Add("Sec-Websocket-Protocol", "graphql-ws")
	c, _, err := websocket.DefaultDialer.Dial(url, headers)

	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Session) ReadOp() (OperationMessage, error) {
	var msg OperationMessage
	err := s.Ws.ReadJSON(&msg)
	if err != nil {
		return OperationMessage{}, err
	}
	return msg, err
}

func (s *Session) Subscribe(query string) (<-chan string, <-chan error) {

	channel := make(chan string)

	s.Ws.WriteJSON(&OperationMessage{
		Type:    startMsg,
		ID:      "test_1", // Do I need to generate a random ID here
		Payload: json.RawMessage(query),
	})

	go func() {
		for {

			msg, err := s.ReadOp()
			if err != nil {
				s.ErrChan <- err
			}
			rawPayload := json.RawMessage(msg.Payload)
			strPayload := string(rawPayload[:])
			channel <- strPayload
			break
		}
		close(channel)
		// close(s.errChan)
	}()

	return channel, s.ErrChan
}

// func main() {

// 	c := wsConnect("ws://localhost:4000/graphql")
// 	defer c.Close()

// 	session := &Session{
// 		ws: c,
// 	}
// 	session.ws.WriteJSON(&operationMessage{Type: connectionInitMsg, Payload: json.RawMessage("{\"Authorization\": \"hello world\"}")})
// 	msg, _ := session.ReadOp()
// 	log.Println(msg.Type)

// 	query := string(`{"query": "subscription { postPublished (topic: \"hello foo\") { id } }"}`)
// 	subscriptionFuture, _ := session.Subscribe(query)
// 	for subscription := range subscriptionFuture {
// 		fmt.Println(subscription)
// 	}

// }
