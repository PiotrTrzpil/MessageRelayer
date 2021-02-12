package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MessageType int

const (
	StartNewRound MessageType = 1 << iota
	ReceivedAnswer
)

type Message struct {
	Type MessageType
	Data []byte
}

func (messageType MessageType) String() string {
	types := make([]string, 0)
	if messageType&StartNewRound != 0 {
		types = append(types, "StartNewRound")
	}
	if messageType&ReceivedAnswer != 0 {
		types = append(types, "ReceivedAnswer")
	}
	return strings.Join(types, "|")
}

func (message *Message) String() string {
	return fmt.Sprintf("%v:%v", message.Type.String(), string(message.Data))
}

func NewMessageOfType(tpe MessageType) Message {
	return Message{
		Type: tpe,
		Data: []byte(strconv.Itoa(0)),
	}
}

func NewMessage(tpe MessageType, data int) Message {
	return Message{
		Type: tpe,
		Data: []byte(strconv.Itoa(data)),
	}
}

func NewArrayOfMessageTypes(types ...MessageType) []Message {
	messages := make([]Message, 0, len(types))
	for _, tpe := range types {
		messages = append(messages, NewMessage(tpe, 0))
	}
	return messages
}

func NewArrayOfMessages(messages ...Message) []Message {
	return messages
}

// ==========================================================================
// MostRecentCache

type MostRecentCache struct {
	buffer []Message
	limit  int
}

func NewMostRecentCache(limit int) *MostRecentCache {
	return &MostRecentCache{
		buffer: make([]Message, 0, limit+1),
		limit:  limit,
	}
}

func (cache *MostRecentCache) Add(message Message) {
	cache.buffer = append(cache.buffer, message)
	if len(cache.buffer) > cache.limit {
		cache.buffer = cache.buffer[1:]
	}
}

func (cache *MostRecentCache) GetAll() []Message {
	return cache.buffer
}

func (cache *MostRecentCache) Clear() {
	cache.buffer = cache.buffer[:0]
}

// ==========================================================================
// MostRecentKeeper

type MostRecentKeeper struct {
	mostRecentMessages map[MessageType]*MostRecentCache
}

func NewMostRecentKeeper() MostRecentKeeper {
	keeper := MostRecentKeeper{
		mostRecentMessages: make(map[MessageType]*MostRecentCache),
	}
	keeper.mostRecentMessages[StartNewRound] = NewMostRecentCache(2)
	keeper.mostRecentMessages[ReceivedAnswer] = NewMostRecentCache(1)
	return keeper
}

func (keeper *MostRecentKeeper) Clear() {
	for _, value := range keeper.mostRecentMessages {
		value.Clear()
	}
}

func (keeper *MostRecentKeeper) Add(message Message) {
	cache, ok := keeper.mostRecentMessages[message.Type]
	if !ok {
		panic(errors.New(fmt.Sprintf("Can't process message type: %v", message.Type)))
	}
	cache.Add(message)
}

func (keeper *MostRecentKeeper) AddAll(messages []Message) {
	for _, message := range messages {
		keeper.Add(message)
	}
}

func (keeper *MostRecentKeeper) GetAll(tpe MessageType) []Message {
	cache, ok := keeper.mostRecentMessages[tpe]
	if !ok {
		panic(errors.New(fmt.Sprintf("Can't find message type: %v", tpe)))
	}
	return cache.GetAll()
}
