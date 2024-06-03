package data_structures

import (
	"log"
	"sync"
)

type SyncMap[T comparable, P any] struct {
	inner_map map[T]P
	lock      sync.Mutex
}

func (s *SyncMap[T, P]) Load(key T) (P, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	value, exists := s.inner_map[key]
	return value, exists
}

func (s *SyncMap[T, P]) Store(key T, value P) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.inner_map[key] = value
}

func (s *SyncMap[T, P]) Delete(key T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.inner_map, key)
}

func (s *SyncMap[T, P]) Print() {
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Println(s.inner_map)
}

func NewSyncMap[T comparable, P any]() *SyncMap[T, P] {
	return &SyncMap[T, P]{
		inner_map: make(map[T]P),
	}
}
