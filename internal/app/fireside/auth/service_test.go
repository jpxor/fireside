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
package auth

import (
	"errors"
	"github.com/jpxor/fireside/internal/app/fireside/user"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockUserService struct {
	ID   string
	Name string
	Hash string
}

func (mus *MockUserService) New(name, hash string) error {
	return nil
}

func (mus *MockUserService) Delete(id string) error {
	return nil
}

func (mus *MockUserService) Get(id string) (*user.Model, error) {
	if id == mus.ID {
		return &user.Model{
			ID:   id,
			Hash: mus.Hash,
			Name: mus.Name,
		}, nil
	} else {
		return &user.Model{
			ID:   id,
			Hash: "wrong",
			Name: "other",
		}, errors.New("no user with that id")
	}
}

func (mus *MockUserService) Update(id string, newModel *user.Model) error {
	return nil
}

func (mus *MockUserService) ForEach(cb user.ForEachCallback) error {
	return nil
}

func TestAuthentication(t *testing.T) {
	var mus = MockUserService{}
	var username = "username"
	var password = "password"
	var uid = "uid"
	var expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJ1aWQiLCJleHAiOjE2MzQ3MTMxNzQsImlzcyI6ImZpcmVzaWRlIGxvY2FsIHByaXZhdGUgYXBwIHNlcnZlciIsInN1YiI6InVzZXJuYW1lIn0.B0nd1Xu5jhRJbLz8vMXgyr65njZ4rselLRDkEJPHJFk"

	// test good path
	auth := New(&mus)
	require.NotNil(t, auth)

	// set private key for tests
	auth.(*authImpl).privateKey = []byte("static-p-key")

	hash, err := auth.Hash(password)
	require.NoError(t, err)
	require.NotNil(t, hash)

	mus.Hash = hash
	mus.Name = username
	mus.ID = uid

	token, err := auth.Authenticate(uid, password)
	require.NoError(t, err)
	require.NotNil(t, token)

	token, err = auth.Refresh(token)
	require.NoError(t, err)
	require.NotNil(t, token)

	// test wrong uid
	token, usrerr := auth.Authenticate("not uid", password)
	require.Error(t, usrerr)
	require.Equal(t, "", token)

	// test wrong password
	token, hasherr := auth.Authenticate(uid, "not password")
	require.Error(t, hasherr)
	require.Equal(t, "", token)

	// no difference in errs
	require.Equal(t, hasherr, usrerr)

	// bad token
	token, err = auth.Refresh("not token")
	require.Error(t, err)
	require.Equal(t, "", token)

	// expired token
	token, err = auth.Refresh(expiredToken)
	require.Error(t, err)
	require.Equal(t, "", token)
}

func TestAuthorization(t *testing.T) {

}
