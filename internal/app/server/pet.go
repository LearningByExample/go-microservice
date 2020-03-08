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
	petIdExpr    = `^\/pet\/(\d*)$`
	petNotIdExpr = `^\/pet$`
	pathNotValid = "no valid path"
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

func (s petHandler) validPet(pet data.Pet) bool {
	return pet.Name != "" && pet.Race != "" && pet.Mod != ""
}

func (s petHandler) postPetRequest(w http.ResponseWriter, r *http.Request) error {
	if s.petNoIdPathReg.MatchString(r.URL.Path) {
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			pet := data.Pet{}
			if err := decoder.Decode(&pet); err == nil {
				if s.validPet(pet) {
					id := s.data.AddPet(pet.Name, pet.Race, pet.Mod)
					w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
					w.Header().Set(constants.Location, fmt.Sprintf("/pet/%d", id))
					w.WriteHeader(http.StatusOK)
					return nil
				} else {
					return resperr.InvalidResource
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
		return resperr.InvalidResource
	}
}

func (s petHandler) putPetRequest(w http.ResponseWriter, r *http.Request) error {
	change := false
	if id, err := s.petID(r.URL.Path); err == nil {
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			pet := data.Pet{}
			if err := decoder.Decode(&pet); err == nil {
				if s.validPet(pet) {
					if change, err = s.data.UpdatePet(id, pet); err != nil {
						return resperr.NotFound
					} else {
						w.Header().Add(constants.ContentType, constants.ApplicationJsonUtf8)
						if change {
							w.WriteHeader(http.StatusOK)
						} else {
							w.WriteHeader(http.StatusNotModified)
						}
					}
					return nil
				} else {
					return resperr.InvalidResource
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

	if rErr != resperr.None {
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
