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
package docstore

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/rs/xid"
)

type Props map[string]string

type DocStore struct {
	FilePath  string
	Documents map[string]Props
	mutex     sync.RWMutex
}

var ErrForEachStop = errors.New(("ForEachStop"))

// Open creates docstore instance and attempts to load
// data from file if it exists
func Open(filepath string) (DocStore, error) {
	var docstore DocStore
	if _, err := os.Stat(filepath); err != nil {
		docstore.FilePath = filepath
		docstore.Documents = make(map[string]Props)
		err = docstore.Save()
		return docstore, err
	}
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return DocStore{}, err
	}
	err = json.Unmarshal([]byte(data), &docstore)
	return docstore, err
}

// Save will sync docstore with its file
func (ds DocStore) Save() error {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	data, err := json.MarshalIndent(ds, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ds.FilePath, data, 0644)
}

// ForEach document in the store executes the callback
func (ds DocStore) ForEach(callback func(id string, props Props) error) error {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()

	var err error
	for id, props := range ds.Documents {
		if err = callback(id, props); err != nil {
			break
		}
	}
	return err
}

// Insert adds the properties document to docstore and
// returns the id of the newly inserted props
func (ds DocStore) Insert(props Props) string {
	ds.mutex.Lock()
	defer ds.mutex.Unlock()

	id := xid.New().String()
	ds.Documents[id] = props
	return id
}

// Get document by id
func (ds DocStore) Get(id string) Props {
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	return ds.Documents[id]
}
