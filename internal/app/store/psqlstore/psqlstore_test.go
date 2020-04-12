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

var (
	mockErr = errors.New("an error has been produced")
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

	type testCase struct {
		name    string
		prepare func(mock sqlmock.Sqlmock, tt testCase)
		want    int
		err     error
	}

	var cases = []testCase{
		{
			name: "should add correctly",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").
					WillReturnRows(mock.NewRows([]string{"id"}).AddRow(tt.want))
				mock.ExpectCommit()
			},
			want: 1,
			err:  nil,
		},
		{
			name: "should error on tx begin error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin().WillReturnError(tt.err)
			},
			want: 0,
			err:  mockErr,
		},
		{
			name: "should error on query error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").WillReturnError(tt.err)
				mock.ExpectRollback()
			},
			want: 0,
			err:  mockErr,
		},
		{
			name: "should error on commit error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").
					WillReturnRows(mock.NewRows([]string{"id"}).AddRow(12))
				mock.ExpectCommit().WillReturnError(tt.err)
			},
			want: 0,
			err:  mockErr,
		},
		{
			name: "should error on query error with rollback error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectQuery(sqlInsert).WithArgs("name", "race", "mod").
					WillReturnError(tt.err)
				mock.ExpectRollback().WillReturnError(errors.New("error on tx rollback"))
			},
			want: 0,
			err:  mockErr,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ps, mock := initDBMock(t)
			defer ps.Close()
			tt.prepare(mock, tt)
			got, err := ps.AddPet("name", "race", "mod")

			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("error adding pet, got id pet %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(err, tt.err) {
				t.Fatalf("error adding pet want %q, got %q", tt.err, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
			ps.Close()
		})
	}

}

func TestMockPosgreSQLPetStore_GetPet(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(mock sqlmock.Sqlmock, tt testCase)
		want    data.Pet
		err     error
	}

	var cases = []testCase{
		{
			name: "should get row",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				var id int64 = 1
				rows := mock.NewRows([]string{"id", "name", "race", "mod"}).AddRow(id, tt.want.Name, tt.want.Race, tt.want.Mod)
				mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnRows(rows)
			},
			want: data.Pet{
				Id:   1,
				Name: "fuffly",
				Race: "dog",
				Mod:  "happy",
			},
			err: nil,
		},
		{
			name: "should get no row",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnError(sql.ErrNoRows)
			},
			want: data.Pet{},
			err:  store.PetNotFound,
		},
		{
			name: "should error on query error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(1).WillReturnError(tt.err)
			},
			want: data.Pet{},
			err:  mockErr,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ps, mock := initDBMock(t)
			defer ps.Close()
			tt.prepare(mock, tt)
			got, err := ps.GetPet(1)

			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("error getting pet, want id pet %v, got  %v", got, tt.want)
			}
			if err != tt.err {
				t.Fatalf("error getting pet want %q, got %q", tt.err, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}
			ps.Close()
		})
	}

}

func TestMockPosgreSQLPetStore_GetAllPets(t *testing.T) {
	type testCase struct {
		name        string
		prepare     func(mock sqlmock.Sqlmock, tt testCase)
		want        []data.Pet
		err         error
		externalErr bool
	}

	var cases = []testCase{
		{
			name: "should get rows",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				rows := mock.NewRows([]string{"id", "name", "race", "mod"})
				for _, pet := range tt.want {
					rows.AddRow(pet.Id, pet.Name, pet.Race, pet.Mod)
				}
				mock.ExpectQuery(sqlSelectAll).WillReturnRows(rows)
			},
			want: []data.Pet{
				{
					Id:   1,
					Name: "name1",
					Race: "race1",
					Mod:  "mod1",
				},
				{
					Id:   2,
					Name: "name2",
					Race: "race2",
					Mod:  "mod2",
				},
				{
					Id:   3,
					Name: "name3",
					Race: "race3",
					Mod:  "mod3",
				},
			},
			err:         nil,
			externalErr: false,
		},
		{
			name: "should get no rows",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelectAll).WillReturnRows(mock.NewRows([]string{"id", "name", "race", "mod"}))
			},
			want:        []data.Pet{},
			err:         nil,
			externalErr: false,
		},
		{
			name: "should error on query error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelectAll).WillReturnError(tt.err)
			},
			want:        []data.Pet{},
			err:         mockErr,
			externalErr: false,
		},
		{
			name: "should error on scan error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				rows := mock.NewRows([]string{"id", "name", "race", "mod"}).
					AddRow("35pp", "name1", "race1", "mod1")
				mock.ExpectQuery(sqlSelectAll).WillReturnRows(rows)
			},
			want:        nil,
			err:         mockErr,
			externalErr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ps, mock := initDBMock(t)
			defer ps.Close()
			tt.prepare(mock, tt)
			got, err := ps.GetAllPets()

			if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("error getting all pets, got %v pets, want %v pets", got, tt.want)
			}
			if tt.externalErr {
				if err == nil {
					t.Fatal("error getting all pets, want an error, got nil")
				}
			} else {
				if err != tt.err {
					t.Fatalf("error getting all pets, want %q, got %q", tt.err, err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMockPosgreSQLPetStore_DeletePet(t *testing.T) {
	type testCase struct {
		name    string
		prepare func(mock sqlmock.Sqlmock, tt testCase)
		err     error
	}

	var cases = []testCase{
		{
			name: "should delete",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			err: nil,
		},
		{
			name: "should not found when delete not existing pet",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 0))
			},
			err: store.PetNotFound,
		},
		{
			name: "should error on tx begin error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin().WillReturnError(tt.err)
			},
			err: mockErr,
		},
		{
			name: "should error on query error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnError(tt.err)
			},
			err: mockErr,
		},
		{
			name: "should error on tx commit error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit().WillReturnError(tt.err)

			},
			err: mockErr,
		},
		{
			name: "should error on rows affected error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewErrorResult(tt.err))
				mock.ExpectRollback()
			},
			err: mockErr,
		},
		{
			name: "should error on rows affected error and rollback error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectBegin()
				mock.ExpectExec(sqlDelete).WithArgs(1).WillReturnResult(sqlmock.NewErrorResult(mockErr))
				mock.ExpectRollback().WillReturnError(errors.New("error in rollback"))
			},
			err: mockErr,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ps, mock := initDBMock(t)
			defer ps.Close()
			tt.prepare(mock, tt)

			got := ps.DeletePet(1)

			if tt.err != got {
				t.Fatalf("Error deleting pet, got %v, want no error", got)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMockPosgreSQLPetStore_UpdatePet(t *testing.T) {
	type testCase struct {
		name        string
		prepare     func(mock sqlmock.Sqlmock, tt testCase)
		want        bool
		err         error
		externalErr bool
	}

	var cases = []testCase{
		{
			name: "should update",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectBegin()
				mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 1))
				mock.ExpectCommit()
			},
			want:        true,
			err:         nil,
			externalErr: false,
		},
		{
			name: "should not update when no changes",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectBegin()
				mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 0))
				mock.ExpectRollback()
			},
			want:        false,
			err:         nil,
			externalErr: false,
		},
		{
			name: "should not found when pet does not exist",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnError(sql.ErrNoRows)
			},
			want:        false,
			err:         store.PetNotFound,
			externalErr: false,
		},
		{
			name: "should error on verify pet error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnError(tt.err)
			},
			want:        false,
			err:         mockErr,
			externalErr: false,
		},
		{
			name: "should error on query error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectBegin()
				mock.ExpectExec(sqlUpdate).WillReturnError(mockErr)
			},
			want:        false,
			err:         mockErr,
			externalErr: false,
		},
		{
			name: "should error on commit error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectBegin()
				mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewResult(5, 1))
				mock.ExpectCommit().WillReturnError(mockErr)
			},
			want:        false,
			err:         mockErr,
			externalErr: false,
		},
		{
			name: "should error on rows affected error",
			prepare: func(mock sqlmock.Sqlmock, tt testCase) {
				mock.ExpectQuery(sqlSelect).WithArgs(5).WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectBegin()
				mock.ExpectExec(sqlUpdate).WillReturnResult(sqlmock.NewErrorResult(tt.err))
				mock.ExpectRollback()
			},
			want:        false,
			err:         mockErr,
			externalErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ps, mock := initDBMock(t)
			defer ps.Close()
			tt.prepare(mock, tt)

			got, err := ps.UpdatePet(5, "name", "race", "mod")
			if err == nil && got != tt.want {
				t.Fatalf("error updating pet, got %t, want %t", got, tt.want)
			}
			if err != tt.err {
				t.Fatalf("error updating pet, got %v, got %v", err, tt.err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPosgreSQLPetStore_Open(t *testing.T) {
	t.Run("we should be able to open a connection", func(t *testing.T) {
		ps := getPetStore(mockFile)
		var mock sqlmock.Sqlmock

		ps.open = func(driverName, dataSourceName string) (db *sql.DB, err error) {
			db, mock, err = sqlmock.New(sqlmock.MonitorPingsOption(true))

			if err == nil && mock != nil {
				//mock.ExpectPing()
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
