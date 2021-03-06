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
	"database/sql"
	"errors"
	"fmt"
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	_ "github.com/lib/pq"
	"log"
	"time"
)

const (
	connectionString = "host=%s port=%d sslmode=%s dbname=%s user=%s password=%s"
	StoreName        = "postgreSQL"
	errRdyQuery      = "error getting value from readiness query"
)

var (
	errReady = errors.New(errRdyQuery)
)

type conFunc func(driverName, dataSourceName string) (*sql.DB, error)

type posgreSQLPetStore struct {
	cfg    config.CfgData
	db     *sql.DB
	logger func(v ...interface{})
	open   conFunc
}

func (p posgreSQLPetStore) IsReady() error {
	var value = 0
	var err error = nil

	if r := p.queryRow(sqlIsReady); r != nil {
		if err = r.Scan(&value); err == nil {
			if value != 1 {
				err = errReady
			}
		}
	}

	return err
}

func (p posgreSQLPetStore) AddPet(name string, race string, mod string) (int, error) {
	var id = 0
	var err error = nil
	var tx *sql.Tx = nil

	if tx, err = p.beginTransaction(); err == nil {
		if r := p.txQueryRow(tx, sqlInsertPet, name, race, mod); r != nil {
			if err = r.Scan(&id); err == nil {
				err = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
		}
	}

	return id, err
}

func (p posgreSQLPetStore) GetPet(id int) (data.Pet, error) {
	var err error = nil
	var pet = data.Pet{}
	if r := p.queryRow(sqlGetPet, id); r != nil {
		err = r.Scan(&pet.Id, &pet.Name, &pet.Race, &pet.Mod)
		if errors.Is(err, sql.ErrNoRows) {
			err = store.PetNotFound
		}
	}
	return pet, err
}

func (p posgreSQLPetStore) GetAllPets() ([]data.Pet, error) {
	var err error = nil
	var pets = make([]data.Pet, 0)
	var r *sql.Rows

	if r, err = p.query(sqlGetAllPets); err == nil {
		//noinspection GoUnhandledErrorResult
		defer r.Close()
		for r.Next() {
			var pet = data.Pet{}
			if err = r.Scan(&pet.Id, &pet.Name, &pet.Race, &pet.Mod); err != nil {
				break
			}
			pets = append(pets, pet)
		}
	}

	return pets, err
}

func (p posgreSQLPetStore) beginTransaction() (*sql.Tx, error) {
	ops := sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	}
	return p.db.BeginTx(context.Background(), &ops)
}

func (p posgreSQLPetStore) DeletePet(id int) error {
	var err error = nil
	var r sql.Result = nil
	var count int64 = 0
	var tx *sql.Tx
	if tx, err = p.beginTransaction(); err == nil {
		if r, err = p.txExec(tx, sqlDeletePet, id); err == nil {
			if count, err = r.RowsAffected(); err == nil {
				if count == 0 {
					err = store.PetNotFound
					_ = tx.Rollback()
				} else {
					err = tx.Commit()
				}
			} else {
				_ = tx.Rollback()
			}
		}
	}

	return err
}

func (p posgreSQLPetStore) verifyPetExists(id int) error {
	var err error = nil
	var petId = 0
	if r := p.queryRow(sqlVerifyPetExists, id); r != nil {
		err = r.Scan(&petId)
		if errors.Is(err, sql.ErrNoRows) {
			err = store.PetNotFound
		}
	}
	return err
}

func (p posgreSQLPetStore) UpdatePet(id int, name string, race string, mod string) (bool, error) {
	var count int64 = 0
	var err error = nil
	var r sql.Result = nil
	var tx *sql.Tx = nil

	if err = p.verifyPetExists(id); err == nil {
		if tx, err = p.beginTransaction(); err == nil {
			if r, err = p.txExec(tx, sqlUpdatePet, id, name, race, mod); err == nil {
				if count, err = r.RowsAffected(); err == nil {
					if count == 0 {
						err = tx.Rollback()
					} else {
						err = tx.Commit()
					}
				} else {
					_ = tx.Rollback()
				}
			}
		}
	}
	return count == 1, err
}

func (p *posgreSQLPetStore) openConnection() (*sql.DB, error) {
	postgreSQLCfg := p.cfg.Store.Postgresql
	connStr := fmt.Sprintf(connectionString,
		postgreSQLCfg.Host,
		postgreSQLCfg.Port,
		postgreSQLCfg.SSLMode,
		postgreSQLCfg.Database,
		postgreSQLCfg.User,
		postgreSQLCfg.Password,
	)
	conn, err := p.open(postgreSQLCfg.Driver, connStr)
	if err != nil && conn != nil {
		conn.SetMaxOpenConns(postgreSQLCfg.Pool.MaxOpenConns)
		conn.SetMaxIdleConns(postgreSQLCfg.Pool.MaxIdleConns)
		conn.SetConnMaxLifetime(time.Duration(postgreSQLCfg.Pool.MaxTimeConns) * time.Millisecond)
	}
	return conn, err
}

func (p posgreSQLPetStore) checkConnection() error {
	return p.db.Ping()
}

func (p posgreSQLPetStore) createTables() error {
	_, err := p.exec(sqlCreateTable)
	return err
}

func (p posgreSQLPetStore) exec(query string, args ...interface{}) (sql.Result, error) {
	p.logger("SQL query:", query, args)
	return p.db.Exec(query, args...)
}

func (p posgreSQLPetStore) txExec(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	p.logger("SQL query:", query, args)
	return tx.Exec(query, args...)
}

func (p posgreSQLPetStore) queryRow(query string, args ...interface{}) *sql.Row {
	p.logger("SQL query:", query, args)
	return p.db.QueryRow(query, args...)
}

func (p posgreSQLPetStore) txQueryRow(tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	p.logger("SQL query:", query, args)
	return tx.QueryRow(query, args...)
}

func (p posgreSQLPetStore) query(query string, args ...interface{}) (*sql.Rows, error) {
	p.logger("SQL query:", query, args)
	return p.db.Query(query, args...)
}

func (p *posgreSQLPetStore) Open() error {
	log.Println("PostgreSQL store opened.")
	var err error = nil

	if p.db, err = p.openConnection(); err == nil {
		if err = p.checkConnection(); err == nil {
			err = p.createTables()
		}
	}

	return err
}

func (p posgreSQLPetStore) logEmpty(_ ...interface{}) {

}

func (p posgreSQLPetStore) Close() error {
	log.Println("PostgreSQL store closed.")
	return p.db.Close()
}

func NewPostgresSQLPetStore(cfg config.CfgData) store.PetStore {
	result := posgreSQLPetStore{
		cfg:    cfg,
		db:     nil,
		logger: log.Println,
		open:   sql.Open,
	}

	if !cfg.Store.Postgresql.LogQueries {
		result.logger = result.logEmpty
	}
	return &result
}
