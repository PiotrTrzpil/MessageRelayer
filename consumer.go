package main

import (
	"fmt"
	"sync"
	"time"
)

type PrinterConsumer struct {
	name            string
	msgType         MessageType
	messagesChannel chan Message
	received        []Message
	processingDelay time.Duration
}

func NewPrinterConsumer(name string, msgType MessageType) *PrinterConsumer {
	channelBufferSize := 2
	return NewPrinterConsumerWithBufferSize(name, msgType, channelBufferSize)
}

func NewPrinterConsumerWithBufferSize(name string, msgType MessageType, channelBufferSize int) *PrinterConsumer {
	return &PrinterConsumer{
		name:            name,
		msgType:         msgType,
		messagesChannel: make(chan Message, channelBufferSize),
		received:        make([]Message, 0),
		processingDelay: 0,
	}
}

func (consumer *PrinterConsumer) Subscribe(relayer MessageRelayer) *PrinterConsumer {
	relayer.SubscribeToMessages(consumer.msgType, consumer.messagesChannel)
	return consumer
}

func (consumer *PrinterConsumer) Start(waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for message := range consumer.messagesChannel {
		consumer.received = append(consumer.received, message)
		fmt.Printf("[%v] Received: %v\n", consumer.name, message.String())
		time.Sleep(consumer.processingDelay)
	}
}

func (consumer *PrinterConsumer) SetProcessingDelay(processingDelay time.Duration) {
	consumer.processingDelay = processingDelay
}
