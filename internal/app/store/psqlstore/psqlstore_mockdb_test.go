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
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	"reflect"
	"testing"
)

func initDBMock(t *testing.T) (*posgreSQLPetStore, sqlmock.Sqlmock) {
	t.Helper()

	ps := getDefaultPetStore()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	ps.db = db
	return ps, mock
}

func addPetResponseWithError(ps *posgreSQLPetStore, t *testing.T, errorWanted error) {
	t.Helper()

	idGot, errorGot := ps.AddPet("name", "race", "mod")

	idWanted := 0
	if idWanted != idGot {
		t.Fatalf("error, wanted petid %d, got petid %d", idWanted, idGot)
	}

	if errorWanted != errorGot {
		t.Fatalf("error wanted %q, error got %q", errorWanted, errorGot)
	}

}

func TestPosgreSQLPetStore_AddPet_BeginFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx begin")
	mock.ExpectBegin().WillReturnError(errorWanted)

	addPetResponseWithError(ps, t, errorWanted)
}

func TestPosgreSQLPetStore_AddPet_QueryFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in query")
	sqlinsert := "INSERT INTO pets .* RETURNING id;"
	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnError(errorWanted)

	addPetResponseWithError(ps, t, errorWanted)
}

func TestPosgreSQLPetStore_AddPet_CommitFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx commit")
	sqlinsert := "INSERT INTO pets .* RETURNING id;"
	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit().WillReturnError(errorWanted)

	addPetResponseWithError(ps, t, errorWanted)
}

func TestPosgreSQLPetStore_AddPet_RollbackFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	sqlinsert := "INSERT INTO pets .* RETURNING id;"
	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnError(errorWanted)
	mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

	addPetResponseWithError(ps, t, errorWanted)
}

func TestPosgreSQLPetStore_AddPet(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	sqlinsert := "INSERT INTO pets .* RETURNING id;"
	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	wanted := 1
	got, err := ps.AddPet("name", "race", "mod")

	if wanted != got {
		t.Fatalf("Wrong pet id, wanted  %d, got  %d", wanted, got)
	}
	if err != nil {
		t.Fatalf("Got an unexpected error %q", err)
	}
}

func getPetResponseWithError(ps *posgreSQLPetStore, t *testing.T, errorWanted error) {
	petGot, errorGot := ps.GetPet(1)

	petWanted := data.Pet{}
	if !reflect.DeepEqual(petGot, petWanted) {
		t.Fatalf("error, wanted pet %v, got pet %v", petGot, petWanted)
	}

	if errorWanted != errorGot {
		t.Fatalf("error wanted %q, error got %q", errorWanted, errorGot)
	}
}

func TestPosgreSQLPetStore_GetPet_QueryFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	sqlselect := "SELECT id, name, race, mod FROM pets WHERE id = .*"
	mock.ExpectQuery(sqlselect).WithArgs(1).WillReturnError(errorWanted)

	getPetResponseWithError(ps, t, errorWanted)
}

func TestPosgreSQLPetStore_GetPet_QueryNoRows(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	sqlselect := "SELECT id, name, race, mod FROM pets WHERE id = .*"
	mock.ExpectQuery(sqlselect).WithArgs(1).WillReturnError(sql.ErrNoRows)

	errorWanted := store.PetNotFound
	getPetResponseWithError(ps, t, errorWanted)
}

func TestPosgreSQLPetStore_GetPet(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	wanted := data.Pet{
		Id:   1,
		Name: "fuffly",
		Race: "dog",
		Mod:  "happy",
	}

	sqlselect := "SELECT id, name, race, mod FROM pets WHERE id = .*"
	var id int64 = 1
	rows := mock.NewRows([]string{"id", "name", "race", "mod"}).AddRow(id, wanted.Name, wanted.Race, wanted.Mod)
	mock.ExpectQuery(sqlselect).WithArgs(1).WillReturnRows(rows)

	got, err := ps.GetPet(1)
	if !reflect.DeepEqual(got, wanted) {
		t.Fatalf("error, wanted pet %v, got pet %v", got, wanted)
	}
	if err != nil {
		t.Fatalf("Got an unexpected error %q", err)
	}
}
