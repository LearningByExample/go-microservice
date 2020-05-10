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
	"context"
	"fmt"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

const (
	postgreSQLFile         = "postgresql.json"
	sqlResetDB             = "DROP TABLE PETS"
	integrationTestSkipped = "Integration test are skipped"
)

var psqlC testcontainers.Container = nil
var host string
var port int

func TestMain(m *testing.M) {
	ctx := context.Background()
	code := m.Run()
	if psqlC != nil {
		defer psqlC.Terminate(ctx)
	}
	os.Exit(code)
}

func getIntegrationPetStore(t *testing.T) *posgreSQLPetStore {
	t.Helper()
	if psqlC == nil {
		var err error = nil
		if psqlC, host, port, err = createContainer(); err != nil {
			t.Fatalf("error starting container %v", err)
		}
	}
	ps := getPetStore(postgreSQLFile)

	ps.cfg.Store.Postgresql.Port = port
	ps.cfg.Store.Postgresql.Host = host

	return ps
}

func createContainer() (testcontainers.Container, string, int, error) {
	var err error = nil
	var container testcontainers.Container = nil

	envs := make(map[string]string)
	envs["POSTGRES_USER"] = "petuser"
	envs["POSTGRES_PASSWORD"] = "petpwd"
	envs["POSTGRES_DB"] = "pets"

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForLog("database system is ready").
			WithOccurrence(2).WithStartupTimeout(60 * time.Second).
			WithPollInterval(5 * time.Second),
		Env: envs,
	}
	container, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	ip := "localhost"
	port := 5432

	if err == nil {
		if ip, err = container.Host(ctx); err == nil {
			var natPort nat.Port
			if natPort, err = container.MappedPort(ctx, "5432"); err == nil {
				port = natPort.Int()
			}
		}
	}

	return container, ip, port, err
}

func runTestSQL(sql string, t *testing.T) {
	ps := getIntegrationPetStore(t)
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

	_, _ = ps.exec(sql)
}

func resetDB(t *testing.T) {
	runTestSQL(sqlResetDB, t)
}

func TestPSqlPetStore_OpenClose(t *testing.T) {
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	t.Run("should work", func(t *testing.T) {
		ps := getIntegrationPetStore(t)

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

func TestPSqlPetStore_IsReady(t *testing.T) {
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	ps := getIntegrationPetStore(t)
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

	got := ps.IsReady()

	if got != nil {
		t.Fatalf("error calling is ready got %v, want nil", got)
	}
}

func TestPSqlPetStore_AddPet(t *testing.T) {
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	ps := getIntegrationPetStore(t)
	resetDB(t)
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
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	ps := getIntegrationPetStore(t)
	resetDB(t)
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
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	ps := getIntegrationPetStore(t)
	resetDB(t)
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
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	ps := getIntegrationPetStore(t)
	resetDB(t)
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
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	ps := getIntegrationPetStore(t)
	resetDB(t)
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

func TestPosgreSQLPetStore_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip(integrationTestSkipped)
	}

	ps := getIntegrationPetStore(t)
	resetDB(t)
	_ = ps.Open()
	//noinspection GoUnhandledErrorResult
	defer ps.Close()

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
}
