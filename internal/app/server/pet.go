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

package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/LearningByExample/go-microservice/internal/app/constants"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/resperr"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request) error
type methodsMap map[string]handlerFunc

type petHandler struct {
	petIdPathReg   *regexp.Regexp
	petNoIdPathReg *regexp.Regexp
	data           store.PetStore
	methods        methodsMap
}

const (
	petIdExpr       = `^\/pets\/(\d*)$`
	petNotIdExpr    = `^\/pets$`
	petLocation     = "/pets/%d"
	pathNotValid    = "no valid path"
	petNameNotEmpty = "pet name cannot be empty"
	petRaceNotEmpty = "pet race cannot be empty"
	petModNotEmpty  = "pet mod cannot be empty"
)

var (
	ErrPathNotValid = errors.New(pathNotValid)
)

func (s petHandler) petID(path string) (int, error) {
	matches := s.petIdPathReg.FindStringSubmatch(path)
	if len(matches) == 2 {
		return strconv.Atoi(matches[1])
	}
	return 0, ErrPathNotValid
}

func (s petHandler) getPetRequest(w http.ResponseWriter, r *http.Request) error {
	if s.petNoIdPathReg.MatchString(r.URL.Path) {
		pets, err := s.data.GetAllPets()
		if err == nil {
			w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
			w.WriteHeader(http.StatusOK)
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(pets); err != nil {
				return resperr.WrittenJson
			}
		}
		return err
	} else {
		if id, err := s.petID(r.URL.Path); err == nil {
			if pet, err := s.data.GetPet(id); err == store.PetNotFound {
				return resperr.NotFound
			} else {
				w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
				w.WriteHeader(http.StatusOK)
				encoder := json.NewEncoder(w)
				if err = encoder.Encode(pet); err != nil {
					return resperr.WrittenJson
				}
				return nil
			}
		} else {
			return resperr.InvalidUrl
		}
	}
}

func (s petHandler) validPet(pet data.Pet) error {
	msg := make([]string, 0, 3)

	if pet.Name == "" {
		msg = append(msg, petNameNotEmpty)
	}
	if pet.Race == "" {
		msg = append(msg, petRaceNotEmpty)
	}
	if pet.Mod == "" {
		msg = append(msg, petModNotEmpty)
	}

	if len(msg) == 0 {
		return nil
	} else {
		return resperr.FromErrorMessage(resperr.InvalidResource, msg)
	}
}

func (s petHandler) postPetRequest(w http.ResponseWriter, r *http.Request) error {
	if s.petNoIdPathReg.MatchString(r.URL.Path) {
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			pet := data.Pet{}
			if err := decoder.Decode(&pet); err == nil {
				if err := s.validPet(pet); err == nil {
					id, err := s.data.AddPet(pet.Name, pet.Race, pet.Mod)
					if err == nil {
						w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
						w.Header().Set(constants.Location, fmt.Sprintf(petLocation, id))
						w.WriteHeader(http.StatusOK)
					}
					return err
				} else {
					return err
				}
			} else {
				return resperr.InvalidResource
			}
		} else {
			return resperr.NotBodyProvided
		}
	} else {
		return resperr.InvalidUrl
	}
}

func (s petHandler) deletePetRequest(w http.ResponseWriter, r *http.Request) error {
	if id, err := s.petID(r.URL.Path); err == nil {
		if err := s.data.DeletePet(id); err == store.PetNotFound {
			return resperr.NotFound
		} else {
			w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
			w.WriteHeader(http.StatusOK)
			return nil
		}
	} else {
		return resperr.InvalidUrl
	}
}

func (s petHandler) putPetRequest(w http.ResponseWriter, r *http.Request) error {
	change := false
	if id, err := s.petID(r.URL.Path); err == nil {
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			pet := data.Pet{}
			if err := decoder.Decode(&pet); err == nil {
				if err := s.validPet(pet); err == nil {
					if change, err = s.data.UpdatePet(id, pet.Name, pet.Race, pet.Mod); err != nil {
						return resperr.NotFound
					} else {
						w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
						if change {
							w.WriteHeader(http.StatusOK)
						} else {
							w.WriteHeader(http.StatusNotModified)
						}
					}
					return err
				} else {
					return err
				}
			} else {
				return resperr.InvalidResource
			}
		} else {
			return resperr.NotBodyProvided
		}
	} else {
		return resperr.InvalidUrl
	}
}

func (s petHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var rErr = resperr.None

	if method, found := s.methods[r.Method]; found {
		if err := method(w, r); err != nil {
			rErr = resperr.FromError(err)
		}
	} else {
		rErr = resperr.BadRequest
	}

	if rErr.Status() != http.StatusOK {
		log.Printf("Error %v in %s request %q", rErr, r.Method, r.URL.Path)
	}

	if rErr.Status() != resperr.None.Status() {
		rErr.Write(w)
	}
}

func (s petHandler) addMethod(httpMethod string, handlerFunc handlerFunc) {
	s.methods[httpMethod] = handlerFunc
}

func NewPetHandler(store store.PetStore) http.Handler {
	ph := petHandler{
		petIdPathReg:   regexp.MustCompile(petIdExpr),
		petNoIdPathReg: regexp.MustCompile(petNotIdExpr),
		data:           store,
		methods:        make(methodsMap),
	}

	ph.addMethod(http.MethodGet, ph.getPetRequest)
	ph.addMethod(http.MethodPost, ph.postPetRequest)
	ph.addMethod(http.MethodDelete, ph.deletePetRequest)
	ph.addMethod(http.MethodPut, ph.putPetRequest)

	return ph
}
