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

const (
	sqlinsert    = "INSERT INTO pets .* RETURNING id;"
	sqlselect    = "SELECT .* FROM pets WHERE .*"
	sqlselectall = "SELECT .* FROM pets ORDER BY .*"
	sqlDelete    = "DELETE FROM pets WHERE .*"
	sqlUpdate    = "UPDATE pets .*"
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

	_, errorGot := ps.AddPet("name", "race", "mod")

	if errorWanted != errorGot {
		t.Fatalf("error wanted %q, error got %q", errorWanted, errorGot)
	}

}

func TestMockPosgreSQLPetStore_AddPet_BeginFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx begin")
	mock.ExpectBegin().WillReturnError(errorWanted)

	addPetResponseWithError(ps, t, errorWanted)
}

func TestMockPosgreSQLPetStore_AddPet_QueryFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in query")
	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnError(errorWanted)
	mock.ExpectRollback()

	addPetResponseWithError(ps, t, errorWanted)
}

func TestMockPosgreSQLPetStore_AddPet_CommitFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx commit")
	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(12))
	mock.ExpectCommit().WillReturnError(errorWanted)

	addPetResponseWithError(ps, t, errorWanted)
}

func TestMockPosgreSQLPetStore_AddPet_RollbackFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnError(errorWanted)
	mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

	addPetResponseWithError(ps, t, errorWanted)
}

func TestMockPosgreSQLPetStore_AddPet(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	mock.ExpectBegin()
	mock.ExpectQuery(sqlinsert).WithArgs("name", "race", "mod").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	wanted := 1
	got, err := ps.AddPet("name", "race", "mod")

	if wanted != got {
		t.Fatalf("Wrong pet id, wanted  %d, got  %d", wanted, got)
	}
	if err != nil {
		t.Fatalf("Error should be nil, but got %q", err)
	}
}

func getPetRunAndCheckResponse(ps *posgreSQLPetStore, t *testing.T, errorWanted error, petWanted data.Pet) {
	petGot, errorGot := ps.GetPet(1)

	if !reflect.DeepEqual(petGot, petWanted) {
		t.Fatalf("error, wanted pet %v, got pet %v", petGot, petWanted)
	}

	if errorWanted != errorGot {
		t.Fatalf("error wanted %q, error got %q", errorWanted, errorGot)
	}
}

func TestMockPosgreSQLPetStore_GetPet_QueryFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectQuery(sqlselect).WithArgs(1).WillReturnError(errorWanted)

	getPetRunAndCheckResponse(ps, t, errorWanted, data.Pet{})
}

func TestMockPosgreSQLPetStore_GetPet_QueryNoRows(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	mock.ExpectQuery(sqlselect).WithArgs(1).WillReturnError(sql.ErrNoRows)

	getPetRunAndCheckResponse(ps, t, store.PetNotFound, data.Pet{})
}

func TestMockPosgreSQLPetStore_GetPet(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	wanted := data.Pet{
		Id:   1,
		Name: "fuffly",
		Race: "dog",
		Mod:  "happy",
	}

	var id int64 = 1
	rows := mock.NewRows([]string{"id", "name", "race", "mod"}).AddRow(id, wanted.Name, wanted.Race, wanted.Mod)
	mock.ExpectQuery(sqlselect).WithArgs(1).WillReturnRows(rows)

	getPetRunAndCheckResponse(ps, t, nil, wanted)
}

func getAllPetsRunAndCheckResponse(ps *posgreSQLPetStore, t *testing.T, errorWanted error, want int) {
	pets, err := ps.GetAllPets()

	got := len(pets)
	if got != want {
		t.Fatalf("Number of pets incorrect, wanted %d pets, got %d pets", got, want)
	}
	if err != errorWanted {
		t.Fatalf("Error wanted %q, but got %q", errorWanted, err)
	}
}

func TestMockPosgreSQLPetStore_GetAllPets(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	rows := mock.NewRows([]string{"id", "name", "race", "mod"}).
		AddRow(1, "name1", "race1", "mod1").
		AddRow(2, "name2", "race2", "mod2").
		AddRow(3, "name2", "race3", "mod3")
	mock.ExpectQuery(sqlselectall).WillReturnRows(rows)

	getAllPetsRunAndCheckResponse(ps, t, nil, 3)
}

func TestMockPosgreSQLPetStore_GetAllPets_Empty(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	rows := mock.NewRows([]string{"id", "name", "race", "mod"})
	mock.ExpectQuery(sqlselectall).WillReturnRows(rows)

	getAllPetsRunAndCheckResponse(ps, t, nil, 0)
}

func TestMockPosgreSQLPetStore_GetAllPets_QueryFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectQuery(sqlselectall).WillReturnError(errorWanted)

	getAllPetsRunAndCheckResponse(ps, t, errorWanted, 0)

}

func TestMockPosgreSQLPetStore_DeletePet(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	mock.ExpectBegin()
	mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := ps.DeletePet(1)

	if err != nil {
		t.Fatalf("Error wanted , but got %q", err)
	}
}

func TestMockPosgreSQLPetStore_DeletePet_NoRowsAffected(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	mock.ExpectBegin()
	mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 0))
	mock.ExpectCommit()

	got := ps.DeletePet(1)

	want := store.PetNotFound
	if got != want {
		t.Fatalf("Error wanted %q, but got %q", want, got)
	}
}

func TestMockPosgreSQLPetStore_DeletePet_BeginFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx begin")
	mock.ExpectBegin().WillReturnError(errorWanted)

	err := ps.DeletePet(1)

	if err != errorWanted {
		t.Fatalf("Error wanted %q, but got %q", errorWanted, err)
	}
}

func TestMockPosgreSQLPetStore_DeletePet_QueryFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectBegin()
	mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnError(errorWanted)
	mock.ExpectRollback()

	err := ps.DeletePet(1)

	if err != errorWanted {
		t.Fatalf("Error wanted %q, but got %q", errorWanted, err)
	}
}

func TestMockPosgreSQLPetStore_DeletePet_CommitFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx commit")
	mock.ExpectBegin()
	mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(errorWanted)

	err := ps.DeletePet(1)

	if err != errorWanted {
		t.Fatalf("Error wanted %q, but got %q", errorWanted, err)
	}
}

func TestMockPosgreSQLPetStore_DeletePet_RollbackFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectBegin()
	mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnError(errorWanted)
	mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

	err := ps.DeletePet(1)

	if err != errorWanted {
		t.Fatalf("Error wanted %q, but got %q", errorWanted, err)
	}
}

func updatePetRubAndCheckResponse(ps *posgreSQLPetStore, t *testing.T, errorWanted error, updated bool) {
	isUpdated, err := ps.UpdatePet(5, "name", "race", "mod")

	if err != errorWanted {
		t.Fatalf("Error wanted %q, but got %q", errorWanted, err)
	}
	if isUpdated != updated {
		t.Fatalf("Update pet is true, expected to be %t", isUpdated)
	}
}

func TestMockPosgreSQLPetStore_UpdatePet_PetDoesNotExist(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	mock.ExpectQuery(sqlselect).WithArgs(5).WillReturnError(sql.ErrNoRows)

	updatePetRubAndCheckResponse(ps, t, store.PetNotFound, false)
}

func TestMockPosgreSQLPetStore_UpdatePet_VerifyPetFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectQuery(sqlselect).WithArgs(5).WillReturnError(errorWanted)

	updatePetRubAndCheckResponse(ps, t, errorWanted, false)
}

func TestMockPosgreSQLPetStore_UpdatePet(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	mock.ExpectQuery(sqlselect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 1))
	mock.ExpectCommit()

	updatePetRubAndCheckResponse(ps, t, nil, true)
}

func TestMockPosgreSQLPetStore_UpdatePet_NoPetChanges(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	mock.ExpectQuery(sqlselect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 0))
	mock.ExpectCommit()

	updatePetRubAndCheckResponse(ps, t, nil, false)
}

func TestMockPosgreSQLPetStore_UpdatePet_QueryFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectQuery(sqlselect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(sqlUpdate).WillReturnError(errorWanted)
	mock.ExpectRollback()

	updatePetRubAndCheckResponse(ps, t, errorWanted, false)
}

func TestMockPosgreSQLPetStore_UpdatePet_CommitFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx commit")
	mock.ExpectQuery(sqlselect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 0))
	mock.ExpectCommit().WillReturnError(errorWanted)

	updatePetRubAndCheckResponse(ps, t, errorWanted, false)
}

func TestMockPosgreSQLPetStore_UpdatePet_RollbackFails(t *testing.T) {
	ps, mock := initDBMock(t)
	defer ps.Close()

	errorWanted := fmt.Errorf("error in tx query")
	mock.ExpectQuery(sqlselect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectBegin()
	mock.ExpectExec(sqlUpdate).WillReturnError(errorWanted)
	mock.ExpectRollback().WillReturnError(fmt.Errorf("error in tx rollback"))

	updatePetRubAndCheckResponse(ps, t, errorWanted, false)
}
