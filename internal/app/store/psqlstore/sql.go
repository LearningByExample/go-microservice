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

const (
	sqlVerify      = "SELECT 1;"
	sqlVerifyPetExists = `
		SELECT
			id
		FROM
			pets
		WHERE
			id = $1;`
	sqlCreateTable = `
		CREATE TABLE IF NOT EXISTS pets (
			id SERIAL PRIMARY KEY,
			name varchar(45) NOT NULL,
			mod varchar(25) NOT NULL,
			race varchar(25) NOT NULL
		);`
	sqlInsertPet = `
		INSERT INTO pets
			(name, race, mod)
		VALUES
			($1, $2, $3)
		RETURNING
			id;`
	sqlGetPet = `
		SELECT
			id, name, race, mod
		FROM
			pets
		WHERE
			id = $1;`
	sqlGetAllPets = `
		SELECT
			id, name, race, mod
		FROM
			pets
		ORDER BY
			id ASC;`
	sqlUpdatePet = `
		UPDATE
			pets
		SET
			name = $2,
			race = $3,
			mod = $4
		WHERE
			id = $1
			AND name <> $2
			AND race <> $3
			AND mod <> $4;`
)
