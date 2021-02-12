package main

import "errors"

type NetworkSocket interface {
	Read() (Message, error)
}

type ArrayNetworkSocket struct {
	messages []Message
	current  int
}

func NewFakeNetworkSocket(messages []Message) *ArrayNetworkSocket {
	return &ArrayNetworkSocket{
		messages: messages,
		current:  0,
	}
}

func (socket *ArrayNetworkSocket) Read() (Message, error) {
	if socket.current >= len(socket.messages) {
		return NewMessage(StartNewRound, 0), errors.New("End of stream")
	}
	message := socket.messages[socket.current]
	socket.current++
	return message, nil
}
