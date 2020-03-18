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
	"github.com/LearningByExample/go-microservice/internal/app/config"
	"github.com/LearningByExample/go-microservice/internal/app/data"
	"github.com/LearningByExample/go-microservice/internal/app/store"
	_ "github.com/lib/pq"
)

const (
	connectionString = "host=%s port=%d sslmode=%s dbname=%s user=%s password=%s"
)

type pSqlPetStore struct {
	cfg config.CfgData
	db  *sql.DB
}

func (p pSqlPetStore) AddPet(name string, race string, mod string) (int, error) {
	panic("implement me")
}

func (p pSqlPetStore) GetPet(id int) (data.Pet, error) {
	panic("implement me")
}

func (p pSqlPetStore) GetAllPets() ([]data.Pet, error) {
	panic("implement me")
}

func (p pSqlPetStore) DeletePet(id int) error {
	panic("implement me")
}

func (p pSqlPetStore) UpdatePet(id int, name string, race string, mod string) (bool, error) {
	panic("implement me")
}

func (p pSqlPetStore) openConnection() (*sql.DB, error) {
	postgreSQLCfg := p.cfg.Store.Postgresql
	connStr := fmt.Sprintf(connectionString,
		postgreSQLCfg.Host,
		postgreSQLCfg.Port,
		postgreSQLCfg.SSLMode,
		postgreSQLCfg.Database,
		postgreSQLCfg.User,
		postgreSQLCfg.Password,
	)
	return sql.Open(postgreSQLCfg.Driver, connStr)
}

func (p pSqlPetStore) checkConnection() error {
	_, err := p.db.Exec(sqlVerify)
	return err
}

func (p pSqlPetStore) createTables() error {
	_, err := p.db.Exec(sqlCreateTable)
	return err
}

func (p *pSqlPetStore) Open() error {
	var err error = nil

	if p.db, err = p.openConnection(); err == nil {
		if err = p.checkConnection(); err == nil {
			err = p.createTables()
		}
	}

	return err
}

func (p pSqlPetStore) Close() error {
	return p.db.Close()
}

func NewPostgresSQLPetStore(cfg config.CfgData) store.PetStore {
	result := pSqlPetStore{
		cfg: cfg,
		db:  nil,
	}
	return &result
}
