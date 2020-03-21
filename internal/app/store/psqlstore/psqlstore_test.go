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

package psqlstore

import (
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	testDataFolder    = "testdata"
	postgreSQLFile    = "postgresql.json"
	postgreSQLBadFile = "postgresql-bad.json"
	sqlResetDB        = "DROP TABLE PETS"
)

func getPetStore(cfgFile string) *posgreSQLPetStore {
	path := filepath.Join(testDataFolder, cfgFile)
	cfg, _ := config.GetConfig(path)
	ps := NewPostgresSQLPetStore(cfg).(*posgreSQLPetStore)
	return ps
}

func getDefaultPetStore() *posgreSQLPetStore {
	return getPetStore(postgreSQLFile)
}

func runTestSQL(sql string) {
	ps := getDefaultPetStore()
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

	_, _ = ps.exec(sql)
}

func resetDB() {
	runTestSQL(sqlResetDB)
}

func TestMain(m *testing.M) {
	resetDB()
	m.Run()
}

func TestNewPostgresSQLPetStore(t *testing.T) {
	cfg := config.CfgData{}
	ps := NewPostgresSQLPetStore(cfg)

	if ps == nil {
		t.Fatalf("want store, got nil")
	}
}

func TestPSqlPetStore_OpenClose(t *testing.T) {
	defer resetDB()
	t.Run("should work", func(t *testing.T) {
		ps := getPetStore(postgreSQLFile)

		err := ps.Open()
		if err != nil {
			t.Fatalf("error on open got %v, want nil", err)
		}

		err = ps.Close()
		if err != nil {
			t.Fatalf("error on close got %v, want nil", err)
		}
	})

	t.Run("should fail", func(t *testing.T) {
		ps := getPetStore(postgreSQLBadFile)

		err := ps.Open()
		if err == nil {
			t.Fatal("error on open got nil, want error")
		}

		err = ps.Close()
		if err != nil {
			t.Fatalf("error on close got %v, want nil", err)
		}
	})
}

func TestPSqlPetStore_AddPet(t *testing.T) {
	defer resetDB()
	ps := getDefaultPetStore()
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

	got, err := ps.AddPet("Fluff", "dog", "happy")

	if err != nil {
		t.Fatalf("error on add pet got %v, want nil", err)
	}

	want := 1
	if got != want {
		t.Fatalf("error inserting pet got %d, want %d", got, want)
	}
}

func petEquals(p data.Pet, name string, race string, mod string) bool {
	return p.Name == name && p.Race == race && p.Mod == mod
}

func TestPosgreSQLPetStore_UpdatePet(t *testing.T) {
	defer resetDB()
	ps := getDefaultPetStore()
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

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
			if err != tt.err {
				t.Fatalf("want err %q, got %q", tt.err, err)
			}

			if got != tt.change {
				t.Fatalf("want %v, got %v", tt.change, got)
			}
			if tt.change {
				pet, _ := ps.GetPet(1)
				if !petEquals(pet, tt.pet.Name, tt.pet.Race, tt.pet.Mod) {
					t.Fatalf("pet was not update correctly")
				}
			}
		})
	}
}

func TestPosgreSQLPetStore_DeletePet(t *testing.T) {
	defer resetDB()
	ps := getDefaultPetStore()
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

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

	_ = ps.Close()
}

func TestPSqlPetStore_GetPet(t *testing.T) {
	defer resetDB()
	ps := getDefaultPetStore()
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

	_, _ = ps.AddPet("Fluff", "dog", "happy")

	t.Run("we should find the pet", func(t *testing.T) {
		got, err := ps.GetPet(1)

		if err != nil {
			t.Fatalf("error on get pet got %v, want nil", err)
		}

		want := data.Pet{
			Id:   1,
			Name: "Fluff",
			Race: "dog",
			Mod:  "happy",
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("error getting pet got %v, want %v", got, want)
		}
	})

	t.Run("we should not find the pet", func(t *testing.T) {
		_, err := ps.GetPet(2)

		if err != store.PetNotFound {
			t.Fatalf("error getting pet got %q, want not found", err)
		}
	})
}

func TestPosgreSQLPetStore_GetAllPets(t *testing.T) {
	defer resetDB()
	ps := getDefaultPetStore()
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

	t.Run("should return empty slice", func(t *testing.T) {
		got, err := ps.GetAllPets()
		want := make([]data.Pet, 0)

		if err != nil {
			t.Fatalf("error on get all pets got %v, want nil", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("error getting all pets got %v, want %v", got, want)
		}
	})

	t.Run("should return two pets", func(t *testing.T) {
		idDog, _ := ps.AddPet("Fluff", "dog", "happy")
		idCat, _ := ps.AddPet("Lion", "cat", "brave")

		got, err := ps.GetAllPets()
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

		if err != nil {
			t.Fatalf("error on get all pets got %v, want nil", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("error getting all pets got %v, want %v", got, want)
		}
	})
}
