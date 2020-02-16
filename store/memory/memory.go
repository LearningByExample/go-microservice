package memory

import (
	"github.com/LearningByExample/go-microservice/data"
	"github.com/LearningByExample/go-microservice/store"
)

type PetMap map[int]data.Pet
type inMemoryPetStore struct {
	pets   PetMap
	lastId int
}

func (s *inMemoryPetStore) DeletePet(id int) error {
	_, err := s.GetPet(id)
	if err != nil {
		return err
	}

	delete(s.pets, id)

	return nil
}

func (s *inMemoryPetStore) AddPet(name string, race string, mod string) int {
	s.lastId++
	id := s.lastId
	s.pets[id] = data.Pet{Id: id, Name: name, Race: race, Mod: mod}
	return id
}

func (s inMemoryPetStore) GetPet(id int) (data.Pet, error) {
	var err error = nil
	value, found := s.pets[id]
	if !found {
		err = store.PetNotFound
	}
	return value, err
}

func NewInMemoryPetStore() store.PetStore {
	var store = inMemoryPetStore{
		pets:   make(PetMap),
		lastId: 0,
	}

	return &store
}
