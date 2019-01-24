package db

import (
	"github.com/ktt-ol/status2/pkg/events"
	"github.com/ktt-ol/status2/pkg/state"
)

type OpenStatePersistence struct {
	dbManager      DbManager
	st             *state.State
	lastOpenStates map[Place]state.OpenValueTs
}

func NewOpenStatePersistence(dbManager DbManager, ev events.EventManager, st *state.State) {
	ops := OpenStatePersistence{dbManager, st, make(map[Place]state.OpenValueTs)}

	for _, openState := range dbManager.GetLastOpenStates() {
		ops.lastOpenStates[openState.Place] = openState.State
	}

	ev.On(events.TOPIC_SPACE_OPEN_STATE, ops.onChange)
	ev.On(events.TOPIC_RADSTELLE_OPEN_STATE, ops.onChange)
	ev.On(events.TOPIC_LAB_3D_OPEN_STATE, ops.onChange)
	ev.On(events.TOPIC_MACHINING_OPEN_STATE, ops.onChange)
}

func (ops *OpenStatePersistence) onChange(topic events.EventName) {
	currentState, _ := ops.st.Open.OpenStateForEvent(topic)

	place, _ := TopicToPlace(topic)
	lastState := ops.lastOpenStates[place]

	if currentState.Value == lastState.Value {
		logger.Debug("The current state is the same in the db. Skipping the INSERT.")
		return
	}

	logger.WithField("topic", topic).Info("Update topic in db")
	ops.dbManager.UpdateOpenState(place, *currentState)
	ops.lastOpenStates[place] = *currentState
}
