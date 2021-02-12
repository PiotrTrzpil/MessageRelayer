package main

import (
	"fmt"
	"sync"
)

func main() {
	fmt.Println("Starting MessageRelayer...")

	messages := NewArrayOfMessageTypes(StartNewRound, StartNewRound, StartNewRound, ReceivedAnswer)
	socket := NewFakeNetworkSocket(messages)
	relayer, _ := NewMessageRelayer(socket, 5)

	var wg sync.WaitGroup
	wg.Add(1)

	consumer1 := NewPrinterConsumer("A", ReceivedAnswer)
	consumer1.Subscribe(relayer)
	go consumer1.Start(&wg)

	relayer.Start()

	wg.Wait()

	fmt.Println("Done!")
}
