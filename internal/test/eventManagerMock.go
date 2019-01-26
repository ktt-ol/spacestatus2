package test

import "github.com/ktt-ol/status2/internal/events"

type EventManagerMock struct {
	OnCount     int
	EmitCount   int
	RemoveCount int
	LastEvent   events.EventName
}

func (em *EventManagerMock) On(topic events.EventName, handler events.EventHandler) events.RegistrationId {
	em.OnCount++
	return events.RegistrationId(0)
}

func (em *EventManagerMock) Emit(topic events.EventName) {
	em.EmitCount++
	em.LastEvent = topic
}

func (em *EventManagerMock) Remove(idToRemove events.RegistrationId) {
	em.RemoveCount++
}
