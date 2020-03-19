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
	"path/filepath"
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
