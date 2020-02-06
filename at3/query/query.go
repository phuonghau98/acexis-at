package query

import "github.com/machinebox/graphql"

func NewMessageMutationRequest() *graphql.Request {
	return graphql.NewRequest(`
	mutation ($content: String!, $roomID: String!) {
		createMessage(message: {
			content: $content,
			roomID: $roomID
		}) {
			_id
			createdBy
			roomID
			content
			createdAt
		}
	}
	`)
}

type Message struct {
	ID        string `json:"_id"`
	CreatedBy string `json:"createdBy"`
	RoomID    string `json:"roomID"`
	Content   string `json:"content"`
	CreatedAt int    `json:"createdAt"`
}

type MessageResponse struct {
	CreateMessage struct {
		Message
	}
}

type MessageSubscriptionResponse struct {
	Data struct {
		MessageCreated Message
	} `json:"data"`
}
