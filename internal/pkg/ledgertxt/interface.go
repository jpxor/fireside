/*
 *  Copyright © 2021 Josh Simonot
 *
 *  fireside is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  fireside is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with fireside. If not, see <https://www.gnu.org/licenses/>.
 */
package ledgertxt

import "os"

// Parser parses entries from ledger-cli
// plain text format
type Parser interface {
	// Next advances the parser to the next entry.
	// Returns io.EOF when there are no more entries.
	Next() (Entry, error)
	// Close will release the file being parsed
	Close()
}

type Entry struct {
	Date     string
	Vendor   string
	Comment  string
	Tags     []string
	Debited  string
	Credited string
	Type     string
	Amount   float64
}

func New(path string) (Parser, error) {
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return NewFileParser(file)
}
