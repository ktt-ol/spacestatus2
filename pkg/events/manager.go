package events

import "sync"

type EventHandler func(topic EventName)

type RegistrationId uint

type EventManager interface {
	On(topic EventName, handler EventHandler) RegistrationId
	Emit(topic EventName)
	Remove(idToRemove RegistrationId)
}

type listEntry struct {
	id      RegistrationId
	handler EventHandler
}

type eventManagerImpl struct {
	idCounter RegistrationId
	listener  map[EventName][]listEntry
	lock      sync.RWMutex
}

func NewEventManager() EventManager {
	instance := eventManagerImpl{listener: make(map[EventName][]listEntry)}

	return &instance
}

func (em *eventManagerImpl) On(topic EventName, handler EventHandler) RegistrationId {
	em.lock.Lock()
	defer em.lock.Unlock()

	em.idCounter++
	entry := listEntry{em.idCounter, handler}
	if handlerList, ok := em.listener[topic]; ok {
		em.listener[topic] = append(handlerList, entry)
	} else {
		em.listener[topic] = []listEntry{entry}
	}

	return em.idCounter
}

func (em *eventManagerImpl) Emit(topic EventName) {
	em.lock.RLock()
	defer em.lock.RUnlock()

	if handlerList, ok := em.listener[topic]; ok {
		for _, listEntry := range handlerList {
			listEntry.handler(topic)
		}
	}
}

func (em *eventManagerImpl) Remove(idToRemove RegistrationId) {
	em.lock.Lock()
	defer em.lock.Unlock()

	for topic, handlerList := range em.listener {
		newHandlerList := make([]listEntry, 0, len(handlerList))
		for _, listEntry := range handlerList {
			if listEntry.id != idToRemove {
				newHandlerList = append(newHandlerList, listEntry)
			}
		}

		em.listener[topic] = newHandlerList
	}
}
