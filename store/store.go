package store

import (
	"errors"
	"github.com/LearningByExample/go-microservice/data"
)

type PetStore interface {
	AddPet(name string, race string, mod string) int
	GetPet(id int) (data.Pet, error)
	DeletePet(id int) error
}

var PetNotFound = errors.New("can not find pet")
