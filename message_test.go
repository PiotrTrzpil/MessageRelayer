package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMostRecentKeeper_preserve_two_StartNewRound(t *testing.T) {
	// given
	keeper := NewMostRecentKeeper()

	// when
	keeper.AddAll(NewArrayOfMessages(
		NewMessage(StartNewRound, 1),
		NewMessage(StartNewRound, 2),
		NewMessage(StartNewRound, 3),
		NewMessage(StartNewRound, 4),
	))
	actual := keeper.GetAll(StartNewRound)

	// then
	require.ElementsMatch(t,
		NewArrayOfMessages(
			NewMessage(StartNewRound, 3),
			NewMessage(StartNewRound, 4),
		),
		actual,
	)
}

func TestMostRecentKeeper_preserve_one_StartNewRound(t *testing.T) {
	// given
	keeper := NewMostRecentKeeper()

	// when
	keeper.AddAll(NewArrayOfMessages(
		NewMessage(ReceivedAnswer, 1),
		NewMessage(ReceivedAnswer, 2),
		NewMessage(ReceivedAnswer, 3),
		NewMessage(ReceivedAnswer, 4),
	))
	actual := keeper.GetAll(ReceivedAnswer)

	// then
	require.ElementsMatch(t,
		NewArrayOfMessages(
			NewMessage(ReceivedAnswer, 4),
		),
		actual,
	)
}
