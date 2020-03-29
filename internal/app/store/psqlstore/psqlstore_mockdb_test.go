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
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	sqlInsert    = "INSERT INTO pets .* RETURNING id;"
	sqlSelect    = "SELECT .* FROM pets WHERE .*"
	sqlSelectAll = "SELECT .* FROM pets ORDER BY .*"
	sqlDelete    = "DELETE FROM pets WHERE .*"
	sqlUpdate    = "UPDATE pets .*"
	mockFile     = "mock.json"
)

func getMockPetStore(cfgFile string) *posgreSQLPetStore {
	path := filepath.Join(testDataFolder, cfgFile)
	cfg, _ := config.GetConfig(path)
	ps := NewPostgresSQLPetStore(cfg).(*posgreSQLPetStore)
	return ps
}

func initDBMock(t *testing.T) (*posgreSQLPetStore, sqlmock.Sqlmock) {
	t.Helper()

	ps := getMockPetStore(mockFile)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	ps.db = db
	return ps, mock
}

func assertAddPet(ps *posgreSQLPetStore, t *testing.T, want error) {
	t.Helper()

	_, err := ps.AddPet("name", "race", "mod")

	if want != err {
		t.Fatalf("error want %q, error got %q", want, err)
	}

}

func assertGetPet(ps *posgreSQLPetStore, t *testing.T, err error, want data.Pet) {
	t.Helper()
	got, gotErr := ps.GetPet(1)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("error, want pet %v, got pet %v", got, want)
	}

	if err != gotErr {
		t.Fatalf("error want %q, error got %q", err, gotErr)
	}
}

func assertGetAllPets(ps *posgreSQLPetStore, t *testing.T, err error, want int) {
	t.Helper()
	pets, gotErr := ps.GetAllPets()

	got := len(pets)
	if got != want {
		t.Fatalf("Number of pets incorrect, want %d pets, got %d pets", got, want)
	}
	if gotErr != err {
		t.Fatalf("Error want %q, but got %q", err, gotErr)
	}
}

func assertUpdatePet(ps *posgreSQLPetStore, t *testing.T, err error, want bool) {
	t.Helper()
	got, gotErr := ps.UpdatePet(5, "name", "race", "mod")

	if gotErr != err {
		t.Fatalf("Error want %q, but got %q", err, gotErr)
	}
	if got != want {
		t.Fatalf("Error in update pet, want %t, got %t", want, got)
	}
}

func TestMockPosgreSQLPetStore_AddPet(t *testing.T) {
	t.Run("should add correctly", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectBegin()
		mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		want := 1
		got, err := ps.AddPet("name", "race", "mod")

		if want != got {
			t.Fatalf("Wrong pet id, want  %d, got  %d", want, got)
		}
		if err != nil {
			t.Fatalf("Error should be nil, but got %q", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on tx begin error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx begin")
		mock.ExpectBegin().WillReturnError(err)

		assertAddPet(ps, t, err)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in query")
		mock.ExpectBegin()
		mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnError(err)
		mock.ExpectRollback()

		assertAddPet(ps, t, err)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on commit error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx commit")
		mock.ExpectBegin()
		mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(12))
		mock.ExpectCommit().WillReturnError(err)

		assertAddPet(ps, t, err)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error with rollback error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx query")
		mock.ExpectBegin()
		mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnError(err)
		mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

		assertAddPet(ps, t, err)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestMockPosgreSQLPetStore_GetPet(t *testing.T) {
	t.Run("should get row", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := data.Pet{
			Id:   1,
			Name: "fuffly",
			Race: "dog",
			Mod:  "happy",
		}

		var id int64 = 1
		rows := mock.NewRows([]string{"id", "name", "race", "mod"}).AddRow(id, want.Name, want.Race, want.Mod)
		mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnRows(rows)

		assertGetPet(ps, t, nil, want)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should get no row", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnError(sql.ErrNoRows)

		assertGetPet(ps, t, store.PetNotFound, data.Pet{})
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx query")
		mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnError(err)

		assertGetPet(ps, t, err, data.Pet{})
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestMockPosgreSQLPetStore_GetAllPets(t *testing.T) {
	t.Run("should get rows", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		rows := mock.NewRows([]string{"id", "name", "race", "mod"}).
			AddRow(1, "name1", "race1", "mod1").
			AddRow(2, "name2", "race2", "mod2").
			AddRow(3, "name2", "race3", "mod3")
		mock.ExpectQuery(sqlSelectAll).WillReturnRows(rows)

		assertGetAllPets(ps, t, nil, 3)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should get no rows", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		rows := mock.NewRows([]string{"id", "name", "race", "mod"})
		mock.ExpectQuery(sqlSelectAll).WillReturnRows(rows)

		assertGetAllPets(ps, t, nil, 0)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx query")
		mock.ExpectQuery(sqlSelectAll).WillReturnError(err)

		assertGetAllPets(ps, t, err, 0)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestMockPosgreSQLPetStore_DeletePet(t *testing.T) {
	t.Run("should delete", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectBegin()
		mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := ps.DeletePet(1)

		if err != nil {
			t.Fatalf("Error want , but got %q", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should not found when delete not existing pet", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectBegin()
		mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 0))

		got := ps.DeletePet(1)

		want := store.PetNotFound
		if got != want {
			t.Fatalf("Error want %q, but got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on tx begin error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := fmt.Errorf("error in tx begin")
		mock.ExpectBegin().WillReturnError(want)

		got := ps.DeletePet(1)

		if got != want {
			t.Fatalf("Error want %q, but got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := fmt.Errorf("error in tx query")
		mock.ExpectBegin()
		mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnError(want)
		mock.ExpectRollback()

		got := ps.DeletePet(1)

		if got != want {
			t.Fatalf("Error want %q, but got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on tx commit error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := fmt.Errorf("error in tx commit")
		mock.ExpectBegin()
		mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(want)

		got := ps.DeletePet(1)

		if got != want {
			t.Fatalf("Error want %q, but got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error with tx rollback error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := fmt.Errorf("error in tx query")
		mock.ExpectBegin()
		mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnError(want)
		mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

		got := ps.DeletePet(1)

		if got != want {
			t.Fatalf("Error want %q, but got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestMockPosgreSQLPetStore_UpdatePet(t *testing.T) {
	t.Run("should update", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectCommit()

		assertUpdatePet(ps, t, nil, true)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should not update when no changes", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 0))
		mock.ExpectCommit()

		assertUpdatePet(ps, t, nil, false)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should not found when pet does not exist", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnError(sql.ErrNoRows)

		assertUpdatePet(ps, t, store.PetNotFound, false)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on verify pet error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx query")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnError(err)

		assertUpdatePet(ps, t, err, false)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx query")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnError(err)
		mock.ExpectRollback()

		assertUpdatePet(ps, t, err, false)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on commit error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx commit")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 0))
		mock.ExpectCommit().WillReturnError(err)

		assertUpdatePet(ps, t, err, false)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error with rollback error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := fmt.Errorf("error in tx query")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnError(err)
		mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

		assertUpdatePet(ps, t, err, false)
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}
