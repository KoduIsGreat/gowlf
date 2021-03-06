package main

import (
	"bytes"
	"database/sql"
	"sort"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var tq = "SELECT distinct fromcomid, tocomid FROM catchment_navigation" +
	" INNER JOIN catchments ON catchments.comid = catchment_navigation.fromcomid" +
	" or catchments.comid = catchment_navigation.tocomid;"
var badQuery = "a very bad query"

func TestPrint(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{
			name: "SimpleToFrom",
			in:   `1,0`,
			want: "0\n1 0\n",
		},
		{
			name: "SimpleFromTo",
			in:   `0,1`,
			want: "0 1\n1\n",
		},
		{
			name: "BasicToFrom",
			in: `1,0
2,1
3,2
`,
			want: "0\n1 0\n2 1\n3 2\n",
		},
		{
			name: "BasicFromTo",
			in: `0,1
1,2
2,3
`,
			want: "0 1\n1 2\n2 3\n3\n",
		},
		{
			name: "TwoPathsToFrom",
			in: `0,1
1,3
3,5
5,6
0,2
2,4
4,6
`,
			want: "\n0 1 2\n1 3\n2 4\n3 5\n4 6\n5 6\n6|\n0 2 1\n1 3\n2 4\n3 5\n4 6\n5 6\n6",
		},
		{
			name: "CyclesToFrom",
			in: `0,1
1,2
2,1
`,
			want: "0\n1 0 2\n2 1\n|0\n1 2 0\n2 1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := mockQuery([]string{"fromcomid", "tocomid"}, tq, tc.in)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			out := bytes.Buffer{}
			catchments, err := fromDb(db, tq)

			if err != nil {
				t.Fatalf("unexpected error %s while creating graph", err)
			}
			catchments.print(&out)
			got := sortByNewLine(out.String())
			want := sortByNewLine(tc.want)
			if strings.Contains(want, "|") {
				split := strings.Split(want, "|")
				var oneMatch bool
				for _, want := range split {
					oneMatch = got == want
				}
				if oneMatch {
					t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
				}
			} else if got != want {
				t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
			}
		})
	}
}

func TestNewGraph(t *testing.T) {
	for _, tc := range []struct {
		name    string
		query   *string
		err     string
		columns []string
		in      string
	}{
		{
			name:    "BadRowType",
			query:   query,
			err:     "error reading row",
			columns: []string{"fromcomid", "tocomid"},
			in: `A,B
C,D
`,
		},
		{
			name:    "BadQuery",
			query:   &badQuery,
			err:     "error with query",
			columns: []string{"fromcomid", "tocomid"},
			in: `0,1
1,2
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := mockQuery(tc.columns, *tc.query, tc.in)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			if _, err := fromDb(db, tq); err == nil {
				t.Fatalf("expected but did not receive fatal error: %s", tc.err)
			}
		})
	}
}

func TestTranspose(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{
			name: "Basic",
			in:   `1,0`,
			want: "0 1\n1\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := mockQuery([]string{"fromcomid", "tocomid"}, tq, tc.in)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			out := bytes.Buffer{}
			catchments, err := fromDb(db, tq)
			if err != nil {
				t.Fatalf("unexpected error %s while creating graph", err)
			}
			catchments = catchments.transpose()
			catchments.print(&out)
			got := sortByNewLine(out.String())
			want := sortByNewLine(tc.want)
			if strings.Contains(want, "|") {
				split := strings.Split(want, "|")
				var oneMatch bool
				for _, want := range split {
					oneMatch = got == want
				}
				if oneMatch {
					t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
				}
			} else if got != want {
				t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
			}
		})
	}
}

func TestSubNetwork(t *testing.T) {
	for _, tc := range []struct {
		name string
		to   int
		in   string
		want string
	}{
		{
			name: "BasicToFrom",
			to:   3,
			in: `1,0
2,1
3,2
4,2
`,
			want: "\n0 1\n1 2\n2 3\n3",
		},
		{
			name: "BasicFromTo",
			to:   1,
			in: `0,1
1,2
2,3
2,4
`,
			want: "\n1 2\n2 3 4\n|\n1 2\n2 4 3\n",
		},
		{
			name: "CyclesBasicToFrom",
			to:   2,
			in: `1,0
2,1
1,2
`,
			want: "0 1\n1 2\n2 1\n",
		},
		{
			name: "CyclesBasicFromTo",
			to:   1,
			in: `0,1
1,2
2,1
`,
			want: "1 2\n2 1\n",
		},
		{
			name: "CyclesWithSplitToFrom",
			to:   2,
			in: `1,0
2,1
1,2
3,2
4,2
`,
			want: "2 1 3 4\n1 2\n|2 3 4 1\n1 2\n|2 4 3 1\n1 2\n|2 4 1 3\n1 2\n",
		},
		{
			name: "CyclesWithSplitFromTo",
			to:   3,
			in: `0,1
1,2
2,1
2,3
2,4
`,
			want: "0 1\n1 2\n2 1 3\n3\n|0 1\n1 2\n2 3 1\n3",
		},
		{
			name: "TwoSplits",
			to:   8,
			in: `0,1
1,2
1,3
2,4
3,4
4,5
4,6
6,7
5,7
7,8
`,
			want: "0 1\n1 2 3\n2 4\n3 4\n4 5 6\n5 7\n6 7\n7 8\n8|0 1\n1 3 2\n2 4\n3 4\n4 5 6\n5 7\n6 7\n7 8\n8|0 1\n1 2 3\n2 4\n3 4\n4 6 5\n5 7\n6 7\n7 8\n8|0 1\n1 3 2\n2 4\n3 4\n4 5 6\n5 7\n6 7\n7 8\n8",
		},
		{
			name: "TwoSplitsMidpoint",
			to:   4,
			in: `0,1
1,2
1,3
2,4
3,4
4,5
4,6
6,7
5,7
7,8
`,
			want: "0 1\n1 2 3\n2 4\n3 4\n4|0 1\n1 3 2\n2 4\n3 4\n4",
		},
		{
			name: "TwoSplitsLeaf",
			to:   5,
			in: `0,1
1,2
1,3
2,4
3,4
4,5
4,6
6,7
5,7
7,8
`,
			want: "0 1\n1 2 3\n2 4\n3 4\n4 5\n5|0 1\n1 3 2\n2 4\n3 4\n4 5\n5",
		},
		{
			name: "TwoSplitsGap",
			to:   10,
			in: `0,1
1,2
1,3
2,4
3,4
4,5
5,6
5,7
6,8
7,9
8,10
9,10
`,
			want: "0 1\n1 2 3\n2 4\n3 4\n4 5\n5 6 7\n6 8\n7 9\n8 10\n9 10\n10|" +
				"0 1\n1 2 3\n2 4\n3 4\n4 5\n5 7 6\n6 8\n7 9\n8 10\n9 10\n10|" +
				"0 1\n1 3 2\n2 4\n3 4\n4 5\n5 6 7\n6 8\n7 9\n8 10\n9 10\n10|" +
				"0 1\n1 3 2\n2 4\n3 4\n4 5\n5 7 6\n6 8\n7 9\n8 10\n9 10\n10",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := mockQuery([]string{"fromcomid", "tocomid"}, tq, tc.in)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			out := bytes.Buffer{}
			catchments, err := fromDb(db, tq)
			if err != nil {
				t.Fatalf("unexpected error %s while creating graph", err)
			}
			subCatchments := catchments.subNetwork(tc.to)
			subCatchments.print(&out)
			got := sortByNewLine(out.String())
			want := sortByNewLine(tc.want)
			if strings.Contains(want, "|") {
				split := strings.Split(want, "|")
				var oneMatch bool
				for _, want := range split {
					oneMatch = got == want
				}
				if oneMatch {
					t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
				}
			} else if got != want {
				t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
			}
		})
	}
}

func sortByNewLine(s string) string {
	sa := strings.Split(s, "\n")
	sort.Strings(sa)
	return strings.Join(sa, "\n")
}

func mockQuery(columns []string, query, rowsCsv string) (*sql.DB, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	mock.ExpectQuery(query).WillReturnRows(mock.NewRows(columns).FromCSVString(rowsCsv))
	return db, nil
}
