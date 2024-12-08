package storage

import (
	"errors"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrKeyExists   = errors.New("key already exists")
	ErrKeyChanged  = errors.New("key changed")
)

type Storage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]string),
		mu:   sync.RWMutex{},
	}
}

func (s *Storage) Create(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if ok {
		return ErrKeyExists
	}
	s.data[key] = value
	return nil
}

func (s *Storage) ValidateCreate(key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	if ok {
		return ErrKeyExists
	}
	return nil
}

func (s *Storage) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

func (s *Storage) ValidateGet(key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}

func (s *Storage) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	s.data[key] = value
	return nil
}

func (s *Storage) ValidateSet(key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}

func (s *Storage) CAS(key, value, oldValue string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	if s.data[key] != oldValue {
		return ErrKeyChanged
	}
	s.data[key] = value
	return nil
}

func (s *Storage) ValidateCAS(key, oldValue string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	if s.data[key] != oldValue {
		return ErrKeyChanged
	}
	return nil
}

func (s *Storage) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	delete(s.data, key)
	return nil
}

func (s *Storage) ValidateDelete(key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	return nil
}
