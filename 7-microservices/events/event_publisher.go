package events

import (
	"stepikGoWebServices/generated/service"
	"sync"
)

type EventSubscribersManager struct {
	id          int
	subscribers map[int]chan *service.Event
	mutex       *sync.RWMutex
}

func NewEventSubscribersManager() *EventSubscribersManager {
	return &EventSubscribersManager{
		subscribers: make(map[int]chan *service.Event),
		mutex:       &sync.RWMutex{},
	}
}

func (s *EventSubscribersManager) Notify(event *service.Event) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, sub := range s.subscribers {
		sub <- event
	}
}

func (s *EventSubscribersManager) NewSub() (int, chan *service.Event) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.id++
	s.subscribers[s.id] = make(chan *service.Event)
	return s.id, s.subscribers[s.id]
}

func (s *EventSubscribersManager) RemoveSub(id int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if subscriber, ok := s.subscribers[id]; ok {
		close(subscriber)
		delete(s.subscribers, id)
	}
}

func (s *EventSubscribersManager) RemoveAll() {
	for id := range s.subscribers {
		s.RemoveSub(id)
	}
}
