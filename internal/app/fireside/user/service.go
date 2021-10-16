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
package user

import (
	"errors"
	"log"

	"github.com/jpxor/fireside/internal/pkg/docstore"
)

type serviceImpl struct {
	DB *docstore.DB
}

// NewService will create user service and
// load any user data from file
func NewService(filepath string) Service {
	db, err := docstore.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	return &serviceImpl{&db}
}

// New creates a new user by name
func (us *serviceImpl) New(name string) error {
	// check if username already in use!
	rc := us.DB.ForEach(func(id string, props docstore.Props) error {
		if props["name"] == name {
			return docstore.ErrForEachStop
		}
		return nil
	})
	if rc == docstore.ErrForEachStop {
		return errors.New(("name conflict"))
	}
	// add new user
	us.DB.Insert(docstore.Props{
		"name": name,
	})
	rc = us.DB.Save()
	if rc != nil {
		return rc
	}
	return nil
}

// ForEach performs action on each user
func (us *serviceImpl) ForEach(callback ForEachCallback) error {
	userCallback := func(id string, props docstore.Props) error {
		return callback(&Model{
			ID:   id,
			Name: props["name"],
		})
	}
	return us.DB.ForEach(userCallback)
}

// Delete
func (us *serviceImpl) Delete(id string) error {
	return errors.New("not implemented")
}

// Get
func (us *serviceImpl) Get(id string) (*Model, error) {
	return nil, errors.New("not implemented")
}

// Update
func (us *serviceImpl) Update(id string, new *Model) error {
	return errors.New("not implemented")
}
