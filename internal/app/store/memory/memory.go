/*
 * Copyright (c) 2020 Learning by Example maintainers.
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in
 *  all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 *  THE SOFTWARE.
 */

package memory

import (
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"sync"
)

type PetMap map[int]data.Pet
type inMemoryPetStore struct {
	pets PetMap
	mu   sync.RWMutex
}

func (s *inMemoryPetStore) DeletePet(id int) error {
	_, err := s.GetPet(id)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pets, id)
	return nil
}

func (s *inMemoryPetStore) AddPet(name string, race string, mod string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := len(s.pets) + 1
	s.pets[id] = data.Pet{Id: id, Name: name, Race: race, Mod: mod}
	return id, nil
}

func (s *inMemoryPetStore) GetPet(id int) (data.Pet, error) {
	var err error = nil

	s.mu.RLock()
	defer s.mu.RUnlock()
	value, found := s.pets[id]

	if !found {
		err = store.PetNotFound
	}
	return value, err
}

func (s *inMemoryPetStore) GetAllPets() ([]data.Pet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]data.Pet, 0, len(s.pets))
	for k := range s.pets {
		result = append(result, s.pets[k])
	}
	return result, nil
}

func petEquals(p data.Pet, name string, race string, mod string) bool {
	return p.Name == name && p.Race == race && p.Mod == mod
}

func (s *inMemoryPetStore) UpdatePet(id int, name string, race string, mod string) (bool, error) {
	var change = false
	found, err := s.GetPet(id)

	if err == nil {
		change = !petEquals(found, name, race, mod)
		if change {
			s.mu.Lock()
			defer s.mu.Unlock()
			s.pets[id] = data.Pet{Id: id, Name: name, Race: race, Mod: mod}
		}
	}

	return change, err
}

func NewInMemoryPetStore() store.PetStore {
	var petStore = inMemoryPetStore{
		pets: make(PetMap),
	}

	return &petStore
}