package db

import (
	"testing"
	"github.com/ktt-ol/status2/pkg/state"
	"github.com/ktt-ol/status2/pkg/events"
	"github.com/stretchr/testify/require"
)

func Test_OpenStatePersistence(t *testing.T) {
	dbMock := new(DbManagerMock)
	dbMock.LastOpenStatesValues = []LastOpenStates{
		{PLACE_SPACE, state.OpenValueTs{state.OPEN, 1234}},
		{PLACE_MACHINING, state.OpenValueTs{state.NONE, 1234}},
	}
	ev := events.NewEventManager()
	appState := state.NewDefaultState()

	NewOpenStatePersistence(dbMock, ev, appState)

	// no changes, because the state was the same
	appState.Open.Space.Value = state.OPEN
	appState.Open.Space.Timestamp = 1
	ev.Emit(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 0, dbMock.UpdateOpenStateCount)

	// new state
	appState.Open.Space.Value = state.NONE
	ev.Emit(events.TOPIC_SPACE_OPEN_STATE)
	require.Equal(t, 1, dbMock.UpdateOpenStateCount)
	require.Equal(t, PLACE_SPACE, dbMock.LastPlace)
	require.Equal(t, state.NONE, dbMock.LastOpenValue.Value)
	require.Equal(t, int64(1), dbMock.LastOpenValue.Timestamp)

	// test another topic
	appState.Open.Machining.Value = state.OPEN
	appState.Open.Machining.Timestamp = 23
	ev.Emit(events.TOPIC_MACHINING_OPEN_STATE)
	require.Equal(t, 2, dbMock.UpdateOpenStateCount)
	require.Equal(t, PLACE_MACHINING, dbMock.LastPlace)
	require.Equal(t, state.OPEN, dbMock.LastOpenValue.Value)
	require.Equal(t, int64(23), dbMock.LastOpenValue.Timestamp)

	// no changes now
	ev.Emit(events.TOPIC_SPACE_OPEN_STATE)
	ev.Emit(events.TOPIC_MACHINING_OPEN_STATE)
	require.Equal(t, 2, dbMock.UpdateOpenStateCount)
}
