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
	"log"
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

func (p *pSqlPetStore) Open() error {
	postgreSQLCfg := p.cfg.Store.Postgresql
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		postgreSQLCfg.Host, postgreSQLCfg.Port, postgreSQLCfg.Database, postgreSQLCfg.User, postgreSQLCfg.Password)
	db, err := sql.Open(postgreSQLCfg.Driver, connStr)
	if err != nil {
		log.Fatal(err)
	}
	p.db = db
	_, err = p.db.Exec(sqlVerify)

	return err
}

func (p *pSqlPetStore) Close() error {
	return p.db.Close()
}

func NewPostgresSQLPetStore(cfg config.CfgData) store.PetStore {
	result := pSqlPetStore{
		cfg: cfg,
		db:  nil,
	}
	return &result
}
