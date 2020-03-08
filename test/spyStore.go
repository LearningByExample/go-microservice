package test

import "github.com/LearningByExample/go-microservice/data"

type SpyStore struct {
	DeleteWasCall bool
	GetWasCall    bool
	AddWasCall    bool
	Id            int
	AddParameters data.Pet
	deleteFunc    func(id int) error
	getFunc       func(id int) (data.Pet, error)
	addFunc       func(name string, race string, mod string) int
}

func (s *SpyStore) Reset() {
	s.DeleteWasCall = false
	s.GetWasCall = false
	s.AddWasCall = false
	s.Id = 0
	s.AddParameters = data.Pet{
		Id:   0,
		Name: "",
		Race: "",
		Mod:  "",
	}
	s.deleteFunc = func(id int) error {
		return nil
	}
	s.getFunc = func(id int) (data.Pet, error) {
		return data.Pet{}, nil
	}
	s.addFunc = func(name string, race string, mod string) int {
		return 0
	}
}

func (s *SpyStore) AddPet(name string, race string, mod string) int {
	s.AddWasCall = true
	s.AddParameters.Name = name
	s.AddParameters.Race = race
	s.AddParameters.Mod = mod
	s.AddParameters.Id = s.addFunc(name, race, mod)
	return s.AddParameters.Id
}

func (s *SpyStore) GetPet(id int) (data.Pet, error) {
	s.GetWasCall = true
	s.Id = id
	return s.getFunc(id)
}

func (s *SpyStore) DeletePet(id int) error {
	s.DeleteWasCall = true
	s.Id = id
	return s.deleteFunc(id)
}

func (s *SpyStore) WhenDeletePet(deleteFunc func(id int) error) {
	s.deleteFunc = deleteFunc
}

func (s *SpyStore) WhenGetPet(getFunc func(id int) (data.Pet, error)) {
	s.getFunc = getFunc
}

func (s *SpyStore) WhenAddPet(addFunc func(name string, race string, mod string) int) {
	s.addFunc = addFunc
}

func NewSpyStore() SpyStore {
	spyStore := SpyStore{}
	spyStore.Reset()
	return spyStore
}
