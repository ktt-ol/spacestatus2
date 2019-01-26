package events

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func Test_Manager(t *testing.T) {
	evManger := NewEventManager()

	handler1Counter := 0
	handler2Counter := 0
	handler3Counter := 0

	id1 := evManger.On(TOPIC_SPACE_DEVICES, func(topic EventName) {
		handler1Counter++
	})
	evManger.On(TOPIC_SPACE_OPEN_STATE, func(topic EventName) {
		handler2Counter++
	})
	evManger.On(TOPIC_POWER_USAGE, func(topic EventName) {
		handler3Counter++
	})

	evManger.Emit(TOPIC_SPACE_DEVICES)
	evManger.Emit(TOPIC_SPACE_DEVICES)
	evManger.Emit(TOPIC_POWER_USAGE)

	require.Equal(t, 2, handler1Counter)
	require.Equal(t, 0, handler2Counter)
	require.Equal(t, 1, handler3Counter)

	evManger.Remove(id1)

	evManger.Emit(TOPIC_SPACE_DEVICES)
	evManger.Emit(TOPIC_POWER_USAGE)

	require.Equal(t, 2, handler1Counter)
	require.Equal(t, 0, handler2Counter)
	require.Equal(t, 2, handler3Counter)
}

