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
	"fmt"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"reflect"
	"sync"
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

	id, _ := ps.AddPet("Fluff", "dog", "happy")

	got, _ := ps.GetPet(id)
	want := data.Pet{
		Id:   id,
		Name: "Fluff",
		Race: "dog",
		Mod:  "happy",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", got, want)
	}
}

func TestAddMultiplePets(t *testing.T) {
	ps := NewInMemoryPetStore()

	_, _ = ps.AddPet("Fluff", "dog", "happy")
	id, _ := ps.AddPet("Lion", "cat", "brave")

	got, _ := ps.GetPet(id)
	want := data.Pet{
		Id:   id,
		Name: "Lion",
		Race: "cat",
		Mod:  "brave",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", got, want)
	}
}

func TestGetNewPetNotFound(t *testing.T) {
	ps := NewInMemoryPetStore()

	_, _ = ps.AddPet("Fluffy", "dog", "happy")

	_, got := ps.GetPet(2)
	want := store.PetNotFound

	if got != want {
		t.Fatalf("want %v, got %v", got, want)
	}
}

func TestDeletePet(t *testing.T) {
	ps := NewInMemoryPetStore()

	_, _ = ps.AddPet("Fluffy", "dog", "happy")

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

func TestUpdatePet(t *testing.T) {
	ps := NewInMemoryPetStore()

	_, _ = ps.AddPet("Fluffy", "dog", "happy")

	type TestCase struct {
		name   string
		id     int
		pet    data.Pet
		change bool
		err    error
	}

	var cases = []TestCase{
		{
			name: "no change pet",
			id:   1,
			pet: data.Pet{
				Name: "Fluffy",
				Race: "dog",
				Mod:  "happy",
			},
			change: false,
			err:    nil,
		},
		{
			name: "change pet",
			id:   1,
			pet: data.Pet{
				Name: "a",
				Race: "b",
				Mod:  "c",
			},
			change: true,
			err:    nil,
		},
		{
			name:   "change not found pet",
			id:     2,
			pet:    data.Pet{},
			change: false,
			err:    store.PetNotFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ps.UpdatePet(tt.id, tt.pet.Name, tt.pet.Race, tt.pet.Mod)
			if got != tt.change {
				t.Fatalf("want %v, got %v", tt.change, got)
			}

			if err != tt.err {
				t.Fatalf("want err %q, got %q", tt.err, err)
			}

			if tt.change {
				pet, _ := ps.GetPet(tt.id)
				if !petEquals(pet, tt.pet.Name, tt.pet.Race, tt.pet.Mod) {
					t.Fatalf("pet was not update correctly")
				}
			}
		})
	}
}

func TestGetPets(t *testing.T) {
	ps := NewInMemoryPetStore()

	idDog, _ := ps.AddPet("Fluff", "dog", "happy")
	idCat, _ := ps.AddPet("Lion", "cat", "brave")

	got, _ := ps.GetAllPets()
	want := []data.Pet{
		{
			Id:   idDog,
			Name: "Fluff",
			Race: "dog",
			Mod:  "happy",
		},
		{
			Id:   idCat,
			Name: "Lion",
			Race: "cat",
			Mod:  "brave",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestConcurrency(t *testing.T) {
	t.Skipf("skiping test, need to fix concurrency")
	ps := NewInMemoryPetStore()
	wantedCount := 1000
	var wg sync.WaitGroup
	wg.Add(wantedCount)
	for i := 0; i < wantedCount; i++ {
		go func(w *sync.WaitGroup) {
			seqName := fmt.Sprintf("Fluff%d", wantedCount)
			id, _ := ps.AddPet(seqName, "dog", "happy")
			_, _ = ps.GetPet(id)
			newName := fmt.Sprintf("Fluffy%d", wantedCount)
			_, _ = ps.UpdatePet(id, newName, "dog", "happy")
			_, _ = ps.GetPet(id)
			_, _ = ps.GetAllPets()
			_ = ps.DeletePet(id)
			_, _ = ps.GetAllPets()
			w.Done()
		}(&wg)
	}

	wg.Wait()
	pets, _ := ps.GetAllPets()
	total := len(pets)
	wantTotal := 0

	if total != wantTotal {
		t.Fatalf("want %q, got %v", wantTotal, total)
	}
}
