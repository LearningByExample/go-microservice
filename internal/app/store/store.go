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

package store

import (
	"errors"
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"log"
)

type PetStore interface {
	AddPet(name string, race string, mod string) (int, error)
	GetPet(id int) (data.Pet, error)
	GetAllPets() ([]data.Pet, error)
	DeletePet(id int) error
	UpdatePet(id int, name string, race string, mod string) (bool, error)
	Open() error
	Close() error
	IsReady() bool
}

type Provider func(cfg config.CfgData) PetStore
type providersMap map[string]Provider

var (
	PetNotFound      = errors.New("can not find pet")
	ProviderNotFound = errors.New("can not find provider")
	providers        = make(providersMap)
)

func AddProvider(name string, provider Provider) {
	log.Printf("Add provider %q.", name)
	providers[name] = provider
}

func GetStoreFromProvider(cfg config.CfgData) (PetStore, error) {
	if provider, found := providers[cfg.Store.Name]; found {
		return provider(cfg), nil
	}
	return nil, ProviderNotFound
}
