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

package _test

import (
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"log"
)

type SpyStore struct {
	DeleteWasCall  bool
	GetWasCall     bool
	GetAllWasCall  bool
	AddWasCall     bool
	UpdateWasCall  bool
	OpenWasCall    bool
	CloseWasCall   bool
	IsReadyWasCall bool
	Id             int
	PetParameters  data.Pet
	deleteFunc     func(id int) error
	getFunc        func(id int) (data.Pet, error)
	getAllFunc     func() ([]data.Pet, error)
	addFunc        func(name string, race string, mod string) (int, error)
	updateFunc     func(id int, name string, race string, mod string) (bool, error)
	openFunc       func() error
	closeFunc      func() error
	isReadyFunc    func() error
}

func (s *SpyStore) Reset() {
	s.DeleteWasCall = false
	s.GetWasCall = false
	s.GetAllWasCall = false
	s.AddWasCall = false
	s.UpdateWasCall = false
	s.OpenWasCall = false
	s.CloseWasCall = false
	s.IsReadyWasCall = false
	s.Id = 0
	s.PetParameters = data.Pet{
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
	s.getAllFunc = func() ([]data.Pet, error) {
		return []data.Pet{}, nil
	}
	s.addFunc = func(name string, race string, mod string) (int, error) {
		return 0, nil
	}
	s.updateFunc = func(id int, name string, race string, mod string) (b bool, err error) {
		return false, nil
	}
	s.openFunc = func() error {
		return nil
	}
	s.closeFunc = func() error {
		return nil
	}
	s.isReadyFunc = func() error {
		return nil
	}
}

func (s *SpyStore) AddPet(name string, race string, mod string) (int, error) {
	var err error = nil
	s.AddWasCall = true
	s.PetParameters.Name = name
	s.PetParameters.Race = race
	s.PetParameters.Mod = mod
	s.PetParameters.Id, err = s.addFunc(name, race, mod)
	return s.PetParameters.Id, err
}

func (s *SpyStore) GetPet(id int) (data.Pet, error) {
	s.GetWasCall = true
	s.Id = id
	return s.getFunc(id)
}

func (s *SpyStore) GetAllPets() ([]data.Pet, error) {
	s.GetAllWasCall = true

	return s.getAllFunc()
}

func (s *SpyStore) DeletePet(id int) error {
	s.DeleteWasCall = true
	s.Id = id
	return s.deleteFunc(id)
}

func (s *SpyStore) UpdatePet(id int, name string, race string, mod string) (bool, error) {
	s.UpdateWasCall = true
	s.Id = id
	s.PetParameters = data.Pet{
		Name: name,
		Race: race,
		Mod:  mod,
	}
	return s.updateFunc(id, name, race, mod)
}

func (s *SpyStore) Open() error {
	s.OpenWasCall = true
	log.Println("Spy pet store opened.")
	return s.openFunc()
}

func (s *SpyStore) Close() error {
	s.CloseWasCall = true
	log.Println("Spy pet store closed.")
	return s.closeFunc()
}

func (s *SpyStore) IsReady() error {
	s.IsReadyWasCall = true
	return s.isReadyFunc()
}

func (s *SpyStore) WhenDeletePet(deleteFunc func(id int) error) {
	s.deleteFunc = deleteFunc
}

func (s *SpyStore) WhenGetPet(getFunc func(id int) (data.Pet, error)) {
	s.getFunc = getFunc
}

func (s *SpyStore) WhenGetAllPets(getAllFunc func() ([]data.Pet, error)) {
	s.getAllFunc = getAllFunc
}

func (s *SpyStore) WhenAddPet(addFunc func(name string, race string, mod string) (int, error)) {
	s.addFunc = addFunc
}

func (s *SpyStore) WhenUpdatePet(updateFunc func(id int, name string, race string, mod string) (bool, error)) {
	s.updateFunc = updateFunc
}

func (s *SpyStore) WhenOpen(openFunc func() error) {
	s.openFunc = openFunc
}

func (s *SpyStore) WhenClose(closeFunc func() error) {
	s.closeFunc = closeFunc
}

func (s *SpyStore) WhenIsReady(isReadyFunc func() error) {
	s.isReadyFunc = isReadyFunc
}

func NewSpyStore() SpyStore {
	spyStore := SpyStore{}
	spyStore.Reset()
	return spyStore
}
