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

package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

var (
	InvalidCfg = errors.New("invalid configuration")
)

type ServerCfg struct {
	Port int `json:"port"`
}

func (cfg ServerCfg) isValid() bool {
	return cfg.Port != 0
}

type StoreCfg struct {
	Name       string        `json:"name"`
	Postgresql PostgreSQLCfg `json:"postgresql"`
}

func (cfg StoreCfg) isValid() bool {
	return cfg.Name != "" && !(cfg.Name == "postgreSQL" && !cfg.Postgresql.isValid())
}

type PoolConfig struct {
	MaxOpenConns int `json:"max-open-conns"`
	MaxIdleConns int `json:"max-idle-conns"`
	MaxTimeConns int `json:"max-time-conns"`
}

type PostgreSQLCfg struct {
	Driver     string     `json:"driver"`
	Host       string     `json:"host"`
	Port       int        `json:"port"`
	SSLMode    string     `json:"ssl-mode"`
	Database   string     `json:"database"`
	User       string     `json:"user"`
	Password   string     `json:"password"`
	LogQueries bool       `json:"log-queries"`
	Pool       PoolConfig `json:"pool"`
}

func (cfg PostgreSQLCfg) isValid() bool {
	return cfg.Driver != "" && cfg.Host != "" && cfg.Port != 0 && cfg.SSLMode != "" &&
		cfg.Database != "" && cfg.User != "" && cfg.Password != "" && cfg.Pool.isValid()
}
func (pool PoolConfig) isValid() bool {
	return pool.MaxOpenConns != 0 && pool.MaxIdleConns != 0 && pool.MaxTimeConns != 0
}

type CfgData struct {
	Server ServerCfg `json:"server"`
	Store  StoreCfg  `json:"store"`
}

func (cfg CfgData) isValid() bool {
	return cfg.Server.isValid() && cfg.Store.isValid()
}

func GetConfig(path string) (CfgData, error) {
	cfg := CfgData{
		Server: ServerCfg{},
		Store:  StoreCfg{},
	}

	file, err := os.Open(path)

	if file != nil && err == nil {
		//noinspection GoUnhandledErrorResult
		defer file.Close()
		var bytes []byte
		bytes, err = ioutil.ReadAll(file)
		if err == nil {
			err = json.Unmarshal(bytes, &cfg)
			if err == nil {
				if !cfg.isValid() {
					err = InvalidCfg
				}
			}
		}
	}

	return cfg, err
}
