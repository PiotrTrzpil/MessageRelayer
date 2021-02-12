package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type MessageRelayer interface {
	SubscribeToMessages(msgType MessageType, messages chan<- Message)
}

type StandardMessageRelayer struct {
	socket            NetworkSocket
	mostRecentKeeper  MostRecentKeeper
	buffer            []Message
	toBroadcast       []Message
	bufferMaxSize     int
	socketMaxWaitTime time.Duration
	readChannel       chan Message
	subscribers       []Subscriber
	subscribersMutex  sync.Mutex
	closed            bool
}

type Subscriber struct {
	msgType     MessageType
	destination chan<- Message
}

func NewMessageRelayer(socket NetworkSocket, bufferSize int) (*StandardMessageRelayer, error) {
	if bufferSize < 1 {
		return nil, errors.New("bufferSize must be equal at least 1")
	}
	return &StandardMessageRelayer{
		socket:            socket,
		mostRecentKeeper:  NewMostRecentKeeper(),
		buffer:            make([]Message, 0, bufferSize),
		toBroadcast:       make([]Message, 0, bufferSize),
		bufferMaxSize:     bufferSize,
		socketMaxWaitTime: 2 * time.Second,
		readChannel:       make(chan Message, bufferSize),
		subscribers:       make([]Subscriber, 0),
	}, nil
}

func (relayer *StandardMessageRelayer) Start() {
	go relayer.ReadFromSocket()
	for {
		err := relayer.FillBuffer()
		toBroadcast := relayer.ProcessBuffer()
		relayer.Broadcast(toBroadcast)
		if err != nil {
			relayer.Close()
			break
		}
	}
}

func (relayer *StandardMessageRelayer) Close() {
	relayer.subscribersMutex.Lock()
	defer relayer.subscribersMutex.Unlock()
	if relayer.closed {
		panic("Relayer is already closed!")
	}
	for _, sub := range relayer.subscribers {
		close(sub.destination)
	}
	relayer.closed = true
}

func (relayer *StandardMessageRelayer) ReadFromSocket() {
	for {
		message, err := relayer.socket.Read()
		if err != nil {
			fmt.Printf("[Relayer] ---- Error when reading from socket: %v\n", err)
			close(relayer.readChannel)
			break
		}
		// send, or block if channel is full
		relayer.readChannel <- message
	}
}

func (relayer *StandardMessageRelayer) FillBuffer() error {
	fmt.Printf("[Relayer] ---- Starting FillBuffer\n")
	relayer.buffer = relayer.buffer[:0]
	for {
		if len(relayer.buffer) < relayer.bufferMaxSize {
			select {
			case message, ok := <-relayer.readChannel:
				if !ok {
					fmt.Printf("[Relayer] No more messages.\n")
					return errors.New("no more messages")
				}
				fmt.Printf("[Relayer] Read message from socket: %v\n", message.String())
				relayer.buffer = append(relayer.buffer, message)
			case <-time.After(relayer.socketMaxWaitTime):
				fmt.Printf("[Relayer] No message received for some time, returning from FillBuffer..\n")
				return nil
			}
		} else {
			fmt.Printf("[Relayer] ---- Buffer full - returning from FillBuffer.\n")
			break
		}
	}
	return nil
}

func (relayer *StandardMessageRelayer) ProcessBuffer() []Message {
	fmt.Printf("[Relayer] Processing buffer of %v messages...\n", len(relayer.buffer))
	relayer.mostRecentKeeper.Clear()
	relayer.mostRecentKeeper.AddAll(relayer.buffer)
	relayer.toBroadcast = relayer.toBroadcast[:0]
	relayer.toBroadcast = append(relayer.toBroadcast, relayer.mostRecentKeeper.GetAll(StartNewRound)...)
	relayer.toBroadcast = append(relayer.toBroadcast, relayer.mostRecentKeeper.GetAll(ReceivedAnswer)...)
	return relayer.toBroadcast
}

func (relayer *StandardMessageRelayer) Broadcast(toSend []Message) {
	fmt.Printf("[Relayer] Starting broadcast of %v messages...\n", len(toSend))
	for _, message := range toSend {
		for index, subscriber := range relayer.subscribers {
			if subscriber.msgType&message.Type != 0 {
				select {
				case subscriber.destination <- message:
					fmt.Printf("[Relayer] Sent message: %v to receiver at index %v \n", message.String(), index)
				default:
					fmt.Printf("[Relayer] Message %v skipped, receiver at index %v is busy\n", message.String(), index)
				}
			}
		}
	}
}

func (relayer *StandardMessageRelayer) SubscribeToMessages(msgType MessageType, messages chan<- Message) {
	relayer.subscribersMutex.Lock()
	defer relayer.subscribersMutex.Unlock()
	if relayer.closed {
		panic("Relayer is already closed!")
	}

	relayer.subscribers = append(relayer.subscribers, Subscriber{
		msgType:     msgType,
		destination: messages,
	})
	fmt.Printf("[Relayer] Added new subscriber for type(s): %v\n", msgType.String())
}
