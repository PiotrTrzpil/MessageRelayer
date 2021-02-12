package main

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

type MockNetworkSocket struct {
	mock.Mock
}

func (socket *MockNetworkSocket) Read() (Message, error) {
	args := socket.Called()
	message := args.Get(0).(Message)
	return message, args.Error(1)
}

func Test_minimal_buffer(t *testing.T) {
	// given
	socket := new(MockNetworkSocket)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(StartNewRound), nil).After(50 * time.Millisecond).Times(3)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), errors.New("End")).After(50 * time.Millisecond).Times(1)

	relayer, err := NewMessageRelayer(socket, 1)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}
	consumer1 := NewPrinterConsumer("S ", StartNewRound).Subscribe(relayer)

	// when
	consumers := NewConsumerGroup(consumer1).Start()
	relayer.Start()
	consumers.Wait()

	// then
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, StartNewRound, StartNewRound), consumer1.received)
}

func Test_small_number_of_messages(t *testing.T) {
	// given
	socket := new(MockNetworkSocket)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(StartNewRound), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), errors.New("End")).After(50 * time.Millisecond).Times(1)

	relayer, err := NewMessageRelayer(socket, 5)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}
	consumer1 := NewPrinterConsumer("SR", StartNewRound|ReceivedAnswer).Subscribe(relayer)

	// when
	consumers := NewConsumerGroup(consumer1).Start()
	relayer.Start()
	consumers.Wait()

	// then
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, ReceivedAnswer), consumer1.received)
}

func Test_messages_just_the_size_of_buffer(t *testing.T) {
	// given
	socket := new(MockNetworkSocket)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(StartNewRound), nil).After(50 * time.Millisecond).Times(3)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), nil).After(50 * time.Millisecond).Times(2)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), errors.New("End")).After(50 * time.Millisecond).Times(1)

	relayer, err := NewMessageRelayer(socket, 5)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}

	consumer1 := NewPrinterConsumer("R ", ReceivedAnswer).Subscribe(relayer)
	consumer2 := NewPrinterConsumer("S ", StartNewRound).Subscribe(relayer)
	consumer3 := NewPrinterConsumer("SR", StartNewRound|ReceivedAnswer).Subscribe(relayer)

	// when
	consumers := NewConsumerGroup(consumer1, consumer2, consumer3).Start()
	relayer.Start()
	consumers.Wait()

	// then
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, StartNewRound), consumer2.received)
	require.Equal(t, NewArrayOfMessageTypes(ReceivedAnswer), consumer1.received)
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, StartNewRound, ReceivedAnswer), consumer3.received)
}

func Test_preserve_latest_messages(t *testing.T) {
	// given
	socket := new(MockNetworkSocket)
	socket.On("Read", mock.Anything).Return(NewMessage(ReceivedAnswer, 1), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessage(ReceivedAnswer, 2), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessage(StartNewRound, 3), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessage(StartNewRound, 4), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), errors.New("End")).After(50 * time.Millisecond).Times(1)

	relayer, err := NewMessageRelayer(socket, 5)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}

	consumer3 := NewPrinterConsumer("SR", StartNewRound|ReceivedAnswer).Subscribe(relayer)

	// when
	consumers := NewConsumerGroup(consumer3).Start()
	relayer.Start()
	consumers.Wait()

	// then
	require.Equal(t, NewArrayOfMessages(
		NewMessage(StartNewRound, 3),
		NewMessage(StartNewRound, 4),
		NewMessage(ReceivedAnswer, 2),
	),
		consumer3.received)
}

func Test_larger_number_of_messages(t *testing.T) {
	// given
	socket := new(MockNetworkSocket)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(StartNewRound), nil).After(50 * time.Millisecond).Times(2)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), nil).After(50 * time.Millisecond).Times(2)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(StartNewRound), nil).After(50 * time.Millisecond).Times(2)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), nil).After(50 * time.Millisecond).Times(8)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), errors.New("End")).After(50 * time.Millisecond).Times(1)

	relayer, err := NewMessageRelayer(socket, 7)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}

	consumer1 := NewPrinterConsumer("R ", ReceivedAnswer).Subscribe(relayer)
	consumer2 := NewPrinterConsumer("S ", StartNewRound).Subscribe(relayer)
	consumer3 := NewPrinterConsumer("SR", StartNewRound|ReceivedAnswer).Subscribe(relayer)

	// when
	consumers := NewConsumerGroup(consumer1, consumer2, consumer3).Start()
	relayer.Start()
	consumers.Wait()

	// then
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, StartNewRound), consumer2.received)
	require.Equal(t, NewArrayOfMessageTypes(ReceivedAnswer, ReceivedAnswer), consumer1.received)
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, StartNewRound, ReceivedAnswer, ReceivedAnswer), consumer3.received)
}

func Test_delay_of_one_consumer(t *testing.T) {
	// given
	socket := new(MockNetworkSocket)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(StartNewRound), nil).After(50 * time.Millisecond).Times(4).After(50 * time.Millisecond)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), errors.New("End")).After(50 * time.Millisecond).Times(1)

	relayer, err := NewMessageRelayer(socket, 2)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}

	consumer1 := NewPrinterConsumerWithBufferSize("SR1", StartNewRound|ReceivedAnswer, 1).Subscribe(relayer)
	consumer2 := NewPrinterConsumerWithBufferSize("SR2", StartNewRound|ReceivedAnswer, 1).Subscribe(relayer)
	consumer2.SetProcessingDelay(2 * time.Second)

	// when
	consumers := NewConsumerGroup(consumer1, consumer2).Start()
	relayer.Start()
	consumers.Wait()

	// then the first consumer should receive all 4 messages
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, StartNewRound, StartNewRound, StartNewRound), consumer1.received)

	// but the slow consumer received only 2
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, StartNewRound), consumer2.received)
}

func Test_delay_of_socket(t *testing.T) {
	// given
	socket := new(MockNetworkSocket)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(StartNewRound), nil).After(50 * time.Millisecond).Times(1)
	socket.On("Read", mock.Anything).Return(NewMessageOfType(ReceivedAnswer), errors.New("End")).After(20 * time.Second).After(50 * time.Millisecond).Times(1)

	relayer, err := NewMessageRelayer(socket, 7)
	if err != nil {
		require.Fail(t, err.Error())
		return
	}

	consumer1 := NewPrinterConsumer("SR1", StartNewRound|ReceivedAnswer).Subscribe(relayer)

	// when
	consumers := NewConsumerGroup(consumer1).Start()
	go relayer.Start()
	consumers.WaitAtMost(3 * time.Second)

	// then some messages should have been received even though the socket is being blocked further
	require.Equal(t, NewArrayOfMessageTypes(StartNewRound, ReceivedAnswer), consumer1.received)
}

// ==================================================================
// ConsumerGroup

type ConsumerGroup struct {
	consumers []*PrinterConsumer
	wg        *sync.WaitGroup
}

func NewConsumerGroup(consumers ...*PrinterConsumer) *ConsumerGroup {
	var wg sync.WaitGroup
	wg.Add(len(consumers))
	return &ConsumerGroup{
		consumers: consumers,
		wg:        &wg,
	}
}

func (group *ConsumerGroup) Start() *ConsumerGroup {
	for _, consumer := range group.consumers {
		go consumer.Start(group.wg)
	}
	return group
}

func (group *ConsumerGroup) Wait() {
	group.wg.Wait()
}

func (group *ConsumerGroup) WaitAtMost(duration time.Duration) {
	ch := make(chan string, 1)
	go func() {
		group.wg.Wait()
		ch <- "done"
	}()
	select {
	case _ = <-ch:
		return
	case <-time.After(duration):
		return
	}
}
