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
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	var testpath = "./users.test.json"
	os.Remove(testpath)

	service := NewService(testpath)
	require.NotNil(t, service)

	err := service.New("Name1")
	require.NoError(t, err)

	err = service.New("Name1")
	require.Error(t, err)

	err = service.ForEach(func(usr *Model) error {
		if usr.Name != "Name1" {
			return errors.New("FAIL")
		}
		return nil
	})
	require.NoError(t, err)

	err = service.ForEach(func(usr *Model) error {
		return errors.New("Test Err")
	})
	require.Error(t, err)
}
