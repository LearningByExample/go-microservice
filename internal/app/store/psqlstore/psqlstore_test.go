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
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"log"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	testDataFolder              = "testdata"
	postgreSQLFileWithoutLogger = "postgresql-no-logger.json"
	postgreSQLBadFile           = "postgresql-bad.json"
	sqlInsert                   = "INSERT INTO pets .* RETURNING id;"
	sqlSelect                   = "SELECT .* FROM pets WHERE .*"
	sqlSelectAll                = "SELECT .* FROM pets ORDER BY .*"
	sqlDelete                   = "DELETE FROM pets WHERE .*"
	sqlUpdate                   = "UPDATE pets .*"
	mockSqlCreateTable          = "CREATE TABLE .*"
	mockFile                    = "mock.json"
)

func getPetStore(cfgFile string) *posgreSQLPetStore {
	path := filepath.Join(testDataFolder, cfgFile)
	cfg, _ := config.GetConfig(path)
	ps := NewPostgresSQLPetStore(cfg).(*posgreSQLPetStore)
	return ps
}

func TestNewPostgresSQLPetStore(t *testing.T) {
	cfg := config.CfgData{}
	ps := NewPostgresSQLPetStore(cfg)

	if ps == nil {
		t.Fatalf("want store, got nil")
	}
}

func TestPSqlPetStore_Logger(t *testing.T) {
	resetDB()
	defer resetDB()
	t.Run("should save logs", func(t *testing.T) {
		ps := getPetStore(postgreSQLFile)
		got := reflect.ValueOf(ps.logger)
		want := reflect.ValueOf(log.Println)
		if got.Pointer() != want.Pointer() {
			t.Fatalf("error getting logger, got %v, want %v", got, want)
		}
	})

	t.Run("should not save logs", func(t *testing.T) {
		ps := getPetStore(postgreSQLFileWithoutLogger)
		got := reflect.ValueOf(ps.logger)
		want := reflect.ValueOf(ps.logEmpty)
		if got.Pointer() != want.Pointer() {
			t.Fatalf("error getting logger, got %v, want %v", got, want)
		}
	})
}

func initDBMock(t *testing.T) (*posgreSQLPetStore, sqlmock.Sqlmock) {
	t.Helper()

	ps := getPetStore(mockFile)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	ps.db = db
	return ps, mock
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

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("error adding pet, want id pet %v, got  %v", got, want)
		}
		if err != nil {
			t.Fatalf("error adding pet want no error, got %q", got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on tx begin error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx begin")
		mock.ExpectBegin().WillReturnError(want)

		_, got := ps.AddPet("name", "race", "mod")

		if want != got {
			t.Fatalf("error adding pet, want error %q, got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in query")
		mock.ExpectBegin()
		mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnError(want)
		mock.ExpectRollback()

		_, got := ps.AddPet("name", "race", "mod")

		if want != got {
			t.Fatalf("error adding pet, want error %q, got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on commit error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx commit")
		mock.ExpectBegin()
		mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(12))
		mock.ExpectCommit().WillReturnError(want)

		_, got := ps.AddPet("name", "race", "mod")

		if want != got {
			t.Fatalf("error adding pet, want error %q, got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error with rollback error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx query")
		mock.ExpectBegin()
		mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnError(want)
		mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

		_, got := ps.AddPet("name", "race", "mod")

		if want != got {
			t.Fatalf("error adding pet, want error %q, got %q", want, got)
		}
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

		got, err := ps.GetPet(1)

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("error getting pet, want id pet %v, got  %v", got, want)
		}
		if err != nil {
			t.Fatalf("error getting pet want no error, got %q", got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should get no row", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnError(sql.ErrNoRows)

		_, got := ps.GetPet(1)

		want := store.PetNotFound
		if want != got {
			t.Fatalf("error getting pet, error want %q, got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx query")
		mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnError(want)

		_, got := ps.GetPet(1)

		if want != got {
			t.Fatalf("error getting pet, error want %q, got %q", want, got)
		}
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

		pets, err := ps.GetAllPets()

		want := 3
		got := len(pets)
		if got != want {
			t.Fatalf("error getting all pets, want %d pets, got %d pets", want, got)
		}
		if err != nil {
			t.Fatalf("error getting all pets, want no error, got %q", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should get no rows", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		rows := mock.NewRows([]string{"id", "name", "race", "mod"})
		mock.ExpectQuery(sqlSelectAll).WillReturnRows(rows)

		_, err := ps.GetAllPets()

		if err != nil {
			t.Fatalf("error getting all pets, want no error, got %q", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		err := errors.New("error in tx query")
		mock.ExpectQuery(sqlSelectAll).WillReturnError(err)

		_, errGot := ps.GetAllPets()

		if err != errGot {
			t.Fatalf("error getting all pets, want %q, got %q", err, errGot)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on scan error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		rows := mock.NewRows([]string{"id", "name", "race", "mod"}).
			AddRow("35pp", "name1", "race1", "mod1")
		mock.ExpectQuery(sqlSelectAll).WillReturnRows(rows)

		_, err := ps.GetAllPets()

		if err == nil {
			t.Fatalf("error getting all pets, want error sql on scan, got not error")
		}

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

		got := ps.DeletePet(1)

		if got != nil {
			t.Fatalf("Error deleting pet, want no error, got %q", got)
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
		if want != got {
			t.Fatalf("Error deleting pet, want %q, got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on tx begin error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx begin")
		mock.ExpectBegin().WillReturnError(want)

		got := ps.DeletePet(1)

		if want != got {
			t.Fatalf("Error deleting pet, want %q, got %q", want, got)
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

		got := ps.DeletePet(1)

		if want != got {
			t.Fatalf("Error deleting pet, want %q, got %q", want, got)
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

		if want != got {
			t.Fatalf("Error deleting pet, want %q, got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on rows affected error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := fmt.Errorf("error in rows affected")
		mock.ExpectBegin()
		mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewErrorResult(want))
		mock.ExpectRollback()

		got := ps.DeletePet(1)

		if want != got {
			t.Fatalf("Error deleting pet, want %q, got %q", want, got)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on rows affected error and rollback error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := fmt.Errorf("error in rows affected")
		mock.ExpectBegin()
		mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewErrorResult(want))
		mock.ExpectRollback().WillReturnError(fmt.Errorf("error in rollback"))

		got := ps.DeletePet(1)

		if want != got {
			t.Fatalf("Error deleting pet, want %q, got %q", want, got)
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

		got, err := ps.UpdatePet(5, "name", "race", "mod")

		if got != true {
			t.Fatalf("error updating pet, want update result %t, got %t", true, got)
		}
		if err != nil {
			t.Fatalf("error updating pet want no error, got %q", err)
		}
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
		mock.ExpectRollback()

		got, err := ps.UpdatePet(5, "name", "race", "mod")

		if got != false {
			t.Fatalf("error updating pet, want update result %t, got %t", false, got)
		}
		if err != nil {
			t.Fatalf("error updating pet want no error, got %q", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should not found when pet does not exist", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnError(sql.ErrNoRows)

		_, err := ps.UpdatePet(5, "name", "race", "mod")

		want := store.PetNotFound
		if want != err {
			t.Fatalf("error updating pet, error want %q, got %q", want, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on verify pet error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx query")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnError(want)

		_, err := ps.UpdatePet(5, "name", "race", "mod")

		if want != err {
			t.Fatalf("error updating pet, error want %q, got %q", want, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on query error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx query")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnError(want)

		_, err := ps.UpdatePet(5, "name", "race", "mod")

		if want != err {
			t.Fatalf("error updating pet, error want %q, got %q", want, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on commit error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error in tx commit")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 1))
		mock.ExpectCommit().WillReturnError(want)

		_, err := ps.UpdatePet(5, "name", "race", "mod")

		if want != err {
			t.Fatalf("error updating pet, error want %q, got %q", want, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("should error on rows affected error", func(t *testing.T) {
		ps, mock := initDBMock(t)
		defer ps.Close()

		want := errors.New("error on tx rows affected")
		mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectBegin()
		mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewErrorResult(want))
		mock.ExpectRollback()

		_, err := ps.UpdatePet(5, "name", "race", "mod")

		if want != err {
			t.Fatalf("error updating pet, error want %q, got %q", want, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestPosgreSQLPetStore_Open(t *testing.T) {
	t.Run("we should be able to open a connection", func(t *testing.T) {
		ps := getPetStore(mockFile)
		var mock sqlmock.Sqlmock

		ps.open = func(driverName, dataSourceName string) (db *sql.DB, err error) {
			db, mock, err = sqlmock.New(sqlmock.MonitorPingsOption(true))

			if err == nil && mock != nil {
				mock.ExpectPing()
				mock.ExpectExec(mockSqlCreateTable).WillReturnResult(sqlmock.NewResult(0, 0))
			}

			return
		}

		err := ps.Open()

		if err != nil {
			t.Fatalf("expect no error got %q", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("we should not been able to open a connection", func(t *testing.T) {
		mockError := errors.New("invalid connection")
		ps := getPetStore(mockFile)

		ps.open = func(driverName, dataSourceName string) (db *sql.DB, err error) {
			return nil, mockError
		}

		err := ps.Open()

		if err != mockError {
			t.Fatalf("invalid error, got %v, want %v", err, mockError)
		}
	})
}
