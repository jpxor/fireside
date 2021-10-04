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

import (
	"io"
	"os"
)

type FileParser struct {
	file *os.File
}

func NewFileParser(file *os.File) (Parser, error) {
	return &FileParser{file}, nil
}

func (parser *FileParser) Next() (Entry, error) {
	return Entry{}, io.EOF
}

func (parser *FileParser) Close() {
	parser.file.Close()
}
