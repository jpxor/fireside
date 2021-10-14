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
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDocStore(t *testing.T) {
	var testpath = "./docstore.test.json"
	os.Remove(testpath)

	var saved_id string
	{
		// creates new docstore
		ds, err := Open(testpath)
		require.NoError(t, err)
		require.Equal(t, ds.FilePath, testpath)

		_, err = os.Stat(testpath)
		require.NoError(t, err)

		// insert document
		id := ds.Insert(Props{
			"test": "value",
		})
		require.NotEqual(t, id, "")
		saved_id = id

		// save document
		err = ds.Save()
		require.NoError(t, err)
	}
	{
		// opens existing docstore
		ds, err := Open(testpath)
		require.NoError(t, err)
		require.Equal(t, ds.FilePath, testpath)

		_, err = os.Stat(testpath)
		require.NoError(t, err)

		count := 0
		ds.ForEach(func(id string, props Props) {
			require.Equal(t, id, saved_id)
			require.Equal(t, props["test"], "value")
			count++
		})
		require.Equal(t, count, 1)

		props := ds.Get(saved_id)
		require.NotNil(t, props)
		require.Equal(t, props["test"], "value")
	}
}
