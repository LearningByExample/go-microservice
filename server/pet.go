package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LearningByExample/go-microservice/data"
	"github.com/LearningByExample/go-microservice/store"
	"net/http"
	"regexp"
	"strconv"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request) int
type methodsMap map[string]handlerFunc

type petHandler struct {
	petIdPathReg   *regexp.Regexp
	petNoIdPathReg *regexp.Regexp
	data           store.PetStore
	methods        methodsMap
}

const contentType = "application/json; charset=utf-8"
const location = "Location"
const pathNotValid = "no valid path"
const petIdExpr = `^\/pet\/(\d*)$`
const petNotIdExpr = `^\/pet\/?`

var ErrPathNotValid = errors.New(pathNotValid)

func (s petHandler) petID(path string) (int, error) {
	matches := s.petIdPathReg.FindStringSubmatch(path)
	if len(matches) == 2 {
		return strconv.Atoi(matches[1])
	}
	return 0, ErrPathNotValid
}

func (s petHandler) getPetRequest(w http.ResponseWriter, r *http.Request) int {
	status := http.StatusBadRequest
	path := r.URL.Path

	id, err := s.petID(path)

	if err == nil {
		pet, err := s.data.GetPet(id)

		if err == store.PetNotFound {
			status = http.StatusNotFound
		} else {
			w.Header().Add("Content-Type", contentType)
			encoder := json.NewEncoder(w)
			err = encoder.Encode(pet)
			if err == nil {
				status = http.StatusOK
			}
		}
	}

	return status
}

func (s petHandler) validPet(pet data.Pet) bool {
	if pet.Name == "" || pet.Race == "" || pet.Mod == "" {
		return false
	}
	return true
}

func (s petHandler) postPetRequest(w http.ResponseWriter, r *http.Request) int {
	status := http.StatusBadRequest

	path := r.URL.Path

	if s.petNoIdPathReg.MatchString(path) {
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			pet := data.Pet{}
			if err := decoder.Decode(&pet); err == nil {
				if s.validPet(pet) {
					id := s.data.AddPet(pet.Name, pet.Race, pet.Mod)
					w.Header().Set(location, fmt.Sprintf("/pet/%d", id))
					status = http.StatusOK
				}
			}
		}
	}

	return status
}

func (s petHandler) deletePetRequest(w http.ResponseWriter, r *http.Request) int {
	status := http.StatusBadRequest
	path := r.URL.Path

	if id, err := s.petID(path); err == nil {
		if err := s.data.DeletePet(id); err == store.PetNotFound {
			status = http.StatusNotFound
		} else {
			status = http.StatusOK
		}
	}

	return status
}

func (s petHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := http.StatusBadRequest

	if method, found := s.methods[r.Method]; found {
		status = method(w, r)
	}

	w.WriteHeader(status)
}

func NewPetHandler(store store.PetStore) http.Handler {
	ph := petHandler{
		petIdPathReg:   regexp.MustCompile(petIdExpr),
		petNoIdPathReg: regexp.MustCompile(petNotIdExpr),
		data:           store,
		methods:        make(methodsMap),
	}

	ph.methods[http.MethodGet] = ph.getPetRequest
	ph.methods[http.MethodPost] = ph.postPetRequest
	ph.methods[http.MethodDelete] = ph.deletePetRequest

	return ph
}
