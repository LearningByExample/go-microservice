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
	"github.com/LearningByExample/go-microservice/data"
	"github.com/LearningByExample/go-microservice/store"
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

const contentType = "Content-Type"
const applicationJsonUtf8 = "application/json; charset=utf-8"
const location = "Location"
const pathNotValid = "no valid path"
const petIdExpr = `^\/pet\/(\d*)$`
const petNotIdExpr = `^\/pet$`
const petNotFound = "pet not found"
const writtenJson = "written json"
const invalidUrl = "invalid url"
const notBodyProvided = "not body provided"
const invalidPet = "invalid pet"
const badRequest = "bad request"

var ErrPathNotValid = errors.New(pathNotValid)

type ErrorResponse struct {
	error
	ErrorStr string `json:"error"`
	status   int
}

var ErrPetNotFound = ErrorResponse{
	ErrorStr: petNotFound,
	status:   http.StatusNotFound,
}
var ErrWrittenJson = ErrorResponse{
	ErrorStr: writtenJson,
	status:   http.StatusInternalServerError,
}
var ErrInvalidUrl = ErrorResponse{
	ErrorStr: invalidUrl,
	status:   http.StatusBadRequest,
}
var ErrInvalidPet = ErrorResponse{
	ErrorStr: invalidPet,
	status:   http.StatusUnprocessableEntity,
}
var ErrNotBodyProvided = ErrorResponse{
	ErrorStr: notBodyProvided,
	status:   http.StatusBadRequest,
}
var ErrBadRequest = ErrorResponse{
	ErrorStr: badRequest,
	status:   http.StatusBadRequest,
}

func (s petHandler) petID(path string) (int, error) {
	matches := s.petIdPathReg.FindStringSubmatch(path)
	if len(matches) == 2 {
		return strconv.Atoi(matches[1])
	}
	return 0, ErrPathNotValid
}

func (s petHandler) getPetRequest(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path
	if id, err := s.petID(path); err == nil {
		if pet, err := s.data.GetPet(id); err == store.PetNotFound {
			return ErrPetNotFound
		} else {
			w.WriteHeader(http.StatusOK)
			encoder := json.NewEncoder(w)
			if err = encoder.Encode(pet); err != nil {
				return ErrWrittenJson
			}
			return nil
		}
	} else {
		return ErrInvalidUrl
	}
}

func (s petHandler) validPet(pet data.Pet) bool {
	return pet.Name != "" && pet.Race != "" && pet.Mod != ""
}

func (s petHandler) postPetRequest(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path

	if s.petNoIdPathReg.MatchString(path) {
		if r.Body != nil {
			decoder := json.NewDecoder(r.Body)
			pet := data.Pet{}
			if err := decoder.Decode(&pet); err == nil {
				if s.validPet(pet) {
					id := s.data.AddPet(pet.Name, pet.Race, pet.Mod)
					w.Header().Set(location, fmt.Sprintf("/pet/%d", id))
					w.WriteHeader(http.StatusOK)
					return nil
				} else {
					return ErrInvalidPet
				}
			} else {
				return ErrInvalidPet
			}
		} else {
			return ErrNotBodyProvided
		}
	} else {
		return ErrInvalidUrl
	}
}

func (s petHandler) deletePetRequest(w http.ResponseWriter, r *http.Request) error {
	path := r.URL.Path

	if id, err := s.petID(path); err == nil {
		if err := s.data.DeletePet(id); err == store.PetNotFound {
			return ErrPetNotFound
		} else {
			w.WriteHeader(http.StatusOK)
			return nil
		}
	} else {
		return ErrInvalidPet
	}
}

func (s petHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(contentType, applicationJsonUtf8)
	if method, found := s.methods[r.Method]; found {
		if err := method(w, r); err != nil {
			encoder := json.NewEncoder(w)
			switch v := err.(type) {
			case ErrorResponse:
				w.WriteHeader(v.status)
				_ = encoder.Encode(err)
			default:
				w.WriteHeader(http.StatusInternalServerError)
				desc := ErrorResponse{ErrorStr: err.Error(),}
				_ = encoder.Encode(desc)
			}
		}
	} else {
		w.WriteHeader(ErrBadRequest.status)
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(ErrBadRequest)
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

	return ph
}
