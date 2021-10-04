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
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

func TestFileParse(t *testing.T) {
	fcontent := `
2015/10/1 Exxon
    // test entry 1: spaces and double slash comment
    Expenses:Auto:Gas         $10.00
    Liabilities:MasterCard   $-10.00

2015/10/2 Exxon
	; test entry 2: tabs and semicolon comment
	Expenses:Auto:Gas         $10.00
	Liabilities:MasterCard   $-10.00

2015/10/3 Exxon
	#test entry 3: hashtag comment
	Expenses:Auto:Gas         $10.00
	Liabilities:MasterCard   $-10.00

2015/10/4 Exxon
    ; test entry 4: #mixed white space & multi #hashtag
	Expenses:Auto:Gas         $10.00
    Liabilities:MasterCard   $-10.00

2015/10/5
	Expenses:Auto:Gas         10 TOK
	Liabilities:MasterCard   10 TOK
`
	expected := []Entry{{
		"2015/10/1",
		"Exxon",
		"test entry 1: spaces and double slash comment",
		[]string{},
		"Liabilities:MasterCard",
		"Expenses:Auto:Gas",
		"$ ",
		10.00,
	}, {
		"2015/10/2",
		"Exxon",
		"test entry 2: tabs and semicolon comment",
		[]string{},
		"Liabilities:MasterCard",
		"Expenses:Auto:Gas",
		"$ ",
		10.00,
	}, {
		"2015/10/3",
		"Exxon",
		"#test entry 3: hashtag comment",
		[]string{"test"},
		"Liabilities:MasterCard",
		"Expenses:Auto:Gas",
		"$ ",
		10.00,
	}, {
		"2015/10/4",
		"Exxon",
		"#test entry 4: #mixed white space & multi #hashtag",
		[]string{"mixed", "hashtag"},
		"Liabilities:MasterCard",
		"Expenses:Auto:Gas",
		"$ ",
		10.00,
	}, {
		"2015/10/5",
		"",
		"#",
		[]string{},
		"Liabilities:MasterCard",
		"Expenses:Auto:Gas",
		" TOK",
		10.00,
	}}
	ftmp, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	require.NotNil(t, ftmp)
	defer os.Remove(ftmp.Name())

	_, err = ftmp.Write([]byte(fcontent))
	require.NoError(t, err)

	parser, err := New(ftmp.Name())
	require.NoError(t, err)
	require.NotNil(t, parser)
	defer parser.Close()

	expected_index := 0
	for {
		entry, err := parser.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.NotNil(t, entry)

		exp_entry := expected[expected_index]
		require.Equal(t, exp_entry.Date, entry.Date)
		require.Equal(t, exp_entry.Vendor, entry.Vendor)
		require.Equal(t, exp_entry.Comment, entry.Comment)
		require.Equal(t, exp_entry.Debited, entry.Debited)
		require.Equal(t, exp_entry.Credited, entry.Credited)
		require.Equal(t, exp_entry.Type, entry.Type)
		require.Equal(t, exp_entry.Amount, entry.Amount)
		require.Equal(t, exp_entry.Tags, entry.Tags)

		expected_index++
	}
	require.Equal(t, len(expected), expected_index)
}
