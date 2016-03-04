package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"isucdc.com/keyescrow/escrow"
	"isucdc.com/keyescrow/server"

	zmq "github.com/pebbe/zmq3"
)

type Client struct {
	Requester *zmq.Socket
}

func NewClient(addr *net.TCPAddr) (*Client, error) {
	req, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("tcp://%v", addr)
	log.Printf("Connecting to %v", url)
	err = req.Connect(url)
	if err != nil {
		return nil, err
	}

	client := &Client{
		Requester: req,
	}

	return client, nil
}

func (c *Client) Dispatch(username string) error {
	req := &server.Dispatch{
		User: username,
	}
	err := req.Send(c.Requester)
	if err != nil {
		return err
	}

	recv, err := c.Requester.RecvBytes(0)
	if err != nil {
		return err
	}

	msg := server.RecvMsg(recv)
	switch msg.(type) {
	case server.Dispatch:
		// Server re-sends dispatch as an ACK
		return nil
	case server.ErrorMessage:
		m := msg.(server.ErrorMessage)
		return errors.New(m.Message)
	default:
		return errors.New("Got unexpected response from server")
	}
}

func (c *Client) GetKey(username string) (*escrow.Key, error) {
	req := &server.KeyRequest{
		User: username,
	}
	err := req.Send(c.Requester)
	if err != nil {
		return nil, err
	}

	recv, err := c.Requester.RecvBytes(0)
	if err != nil {
		return nil, err
	}

	msg := server.RecvMsg(recv)
	switch msg.(type) {
	case server.ErrorMessage:
		m := msg.(server.ErrorMessage)
		return nil, errors.New(fmt.Sprintf("The server returned an error: %v", m.Message))
	case server.KeyResponse:
		m := msg.(server.KeyResponse)
		return &escrow.Key{
			User:       username,
			PublicKey:  m.PubKey,
			PrivateKey: m.PrivKey,
		}, nil
	default:
		return nil, errors.New("Got unexpected response from server")
	}
}

func (c *Client) SetKey(user string, pubkey string, privkey string) error {
	kr := &server.KeyResponse{
		User: user,
	}
	pub, err := os.Open(pubkey)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(pub)
	if err != nil {
		return err
	}
	kr.PubKey = string(data)

	priv, err := os.Open(privkey)
	if err != nil {
		return err
	}
	data, err = ioutil.ReadAll(priv)
	if err != nil {
		return err
	}
	kr.PrivKey = string(data)

	err = kr.Send(c.Requester)
	if err != nil {
		return err
	}

	recv, err := c.Requester.RecvBytes(0)
	if err != nil {
		return err
	}

	msg := server.RecvMsg(recv)
	switch msg.(type) {
	case server.ErrorMessage:
		m := msg.(server.ErrorMessage)
		return errors.New(m.Message)
	case server.KeyRequest:
		log.Println("Successfully set key")
		return nil
	default:
		return errors.New("Got unexpectd response from server")
	}
	return nil
}

func (c *Client) Close() {
	c.Requester.Close()
}
