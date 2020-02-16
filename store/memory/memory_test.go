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
	"github.com/LearningByExample/go-microservice/data"
	"github.com/LearningByExample/go-microservice/store"
	"reflect"
	"testing"
)

func TestNewPetStore(t *testing.T) {

	got := NewInMemoryPetStore()

	if got == nil {
		t.Fatalf("want PetStore, got nil")
	}
}

func TestAddNewPet(t *testing.T) {
	ps := NewInMemoryPetStore()

	id := ps.AddPet("pelusa", "dog", "happy")

	got, _ := ps.GetPet(id)
	want := data.Pet{
		Id:   id,
		Name: "pelusa",
		Race: "dog",
		Mod:  "happy",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", got, want)
	}
}

func TestAddMultiplePets(t *testing.T) {
	ps := NewInMemoryPetStore()

	ps.AddPet("pelusa1", "dog", "happy")
	id := ps.AddPet("pelusa2", "dog", "happy")
	ps.AddPet("pelusa3", "dog", "happy")

	got, _ := ps.GetPet(id)
	want := data.Pet{
		Id:   id,
		Name: "pelusa2",
		Race: "dog",
		Mod:  "happy",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", got, want)
	}
}

func TestGetNewPetNotFound(t *testing.T) {
	ps := NewInMemoryPetStore()

	ps.AddPet("pelusa", "dog", "happy")

	_, got := ps.GetPet(2)
	want := store.PetNotFound

	if got != want {
		t.Fatalf("want %v, got %v", got, want)
	}
}

func TestDeletePet(t *testing.T) {
	ps := NewInMemoryPetStore()

	ps.AddPet("pelusa", "dog", "happy")

	t.Run("we could delete a existing pet", func(t *testing.T) {
		got := ps.DeletePet(1)
		if got != nil {
			t.Fatalf("want nil, got %v", got)
		}
	})

	t.Run("we could not find a deleted pet", func(t *testing.T) {
		_, got := ps.GetPet(1)
		want := store.PetNotFound
		if got != want {
			t.Fatalf("want %v, got %v", want, got)
		}
	})

	t.Run("we could not delete a not existing pet", func(t *testing.T) {
		got := ps.DeletePet(1)
		want := store.PetNotFound
		if got != want {
			t.Fatalf("want %v, got %v", want, got)
		}
	})
}
