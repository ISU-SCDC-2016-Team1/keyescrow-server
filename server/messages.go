package server

import (
	"encoding/json"
	"log"

	zmq "github.com/pebbe/zmq3"
)

// Message Types
const (
	MSG_ERR float64 = iota
	MSG_KEY_REQUEST
	MSG_KEY_RESPONSE
	MSG_KEY_DISPATCH
)

type Message interface {
	Send(socket *zmq.Socket) error
}

func RecvMsg(data []byte) Message {
	// We didn't get any data
	if len(data) == 0 {
		return nil
	}

	var i interface{}
	if err := json.Unmarshal(data, &i); err != nil {
		log.Printf("Error unmarshaling request: %v", err.Error())
		return nil
	}

	m := i.(map[string]interface{})
	switch m["id"].(float64) {
	case MSG_KEY_REQUEST:
		log.Println("Got KEY_REQUEST")
		var kr KeyRequest
		if err := json.Unmarshal(data, &kr); err != nil {
			log.Printf("Error unmarshaling key request: %v", err.Error())
			return nil
		}
		return kr

	case MSG_KEY_RESPONSE:
		log.Println("Got KEY_RESPONSE")
		var kr KeyResponse
		if err := json.Unmarshal(data, &kr); err != nil {
			log.Printf("Error unmarshaling key response: %v", err.Error())
			return nil
		}
		return kr

	case MSG_KEY_DISPATCH:
		log.Println("Got KEY_DISPACH")
		var d Dispatch
		if err := json.Unmarshal(data, &d); err != nil {
			log.Printf("Error unmarshaling dispatch: %v", err.Error())
			return nil
		}
		return d

	case MSG_ERR:
		log.Println("Got ERR")
		var e ErrorMessage
		if err := json.Unmarshal(data, &e); err != nil {
			log.Printf("The server sent an error, additionaly an error was encountered parsing it: %v", err.Error())
			return nil
		}
		return e

	default:
		log.Println("MSG Not Recognized: %d", m["id"])
		return nil
	}
}

type KeyRequest struct {
	ID   float64 `json:"id"`
	User string  `json:"user"`
}

func (kr KeyRequest) Send(socket *zmq.Socket) error {
	kr.ID = MSG_KEY_REQUEST

	b, err := json.Marshal(kr)

	log.Printf("Sending KeyRequest for user: %v", kr.User)
	_, err = socket.SendBytes(b, 0)
	return err
}

type KeyResponse struct {
	ID      float64 `json:"id"`
	User    string  `json:"user"`
	PubKey  string  `json:"pubkey"`
	PrivKey string  `json:"privkey"`
}

func (kr KeyResponse) Send(socket *zmq.Socket) error {
	kr.ID = MSG_KEY_RESPONSE
	b, err := json.Marshal(kr)
	_, err = socket.SendBytes(b, 0)
	return err
}

type ErrorMessage struct {
	ID      float64 `json:"id"`
	Message string  `json:"message"`
}

func (e ErrorMessage) Send(socket *zmq.Socket) error {
	e.ID = MSG_ERR

	b, err := json.Marshal(e)

	log.Printf("Sending Error")
	_, err = socket.SendBytes(b, 0)
	return err
}

type Dispatch struct {
	ID   float64 `json:"id"`
	User string  `json:"user"`
}

func (d Dispatch) Send(socket *zmq.Socket) error {
	d.ID = MSG_KEY_DISPATCH
	b, err := json.Marshal(d)
	_, err = socket.SendBytes(b, 0)
	return err
}
