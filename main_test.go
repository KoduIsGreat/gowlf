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
			name: "Simple",
			in: `0,1`,
			want: `0	1
1	`,
		},
		{
			name: "Basic",
			in: `0,1
1,2
2,3
`,
			want: `0	1
1	2  
2	3  
3	`,
		},
		{
			name: "TwoPaths",
			in: `0,1
1,3
3,5
5,6
0,2
2,4
4,6
`,
			want: `	1 -> 3
	3 -> 5
	5 -> 6
	2 -> 4
	4 -> 6
`,
		},
		{
			name: "Cycles",
			in: `0,1
1,2
2,1
`,
			want: `	1 -> 2
	2 -> 1
`,
		},
		{
			name: "RealisticExample_UnaBasin",
			in: `0,307562200
0,307578700
0,307601400
0,307635600
0,307668700
0,307676500
307562200,307586600
307578700,307586600
307586600,307586700
307586700,307586800
307586800,307592300
307601400,307586800
`,
			want: `	307578700 -> 307586600
	307586600 -> 307586700
	307586700 -> 307586800
	307586800 -> 307592300
	307601400 -> 307586800
	307562200 -> 307586600
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := mockQuery([]string{"fromcomid", "tocomid"}, tq, tc.in)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			out := bytes.Buffer{}
			catchments, err := fromDB(db, tq)

			if err != nil {
				t.Fatalf("unexpected error %s while creating graph", err)
			}
			catchments.sprint(&out)
			got := sortByNewLine(out.String())
			want := sortByNewLine(tc.want)
			if got != want {
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
			name:    "NoRoot",
			query:   &query,
			err:     "root %d does not exist",
			columns: []string{"fromcomid", "tcomid"},
			in: `1,2
2,1
`,
		},
		{
			name:    "BadRowType",
			query:   &query,
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

			if _, err := fromDB(db, tq); err == nil {
				t.Fatalf("expected but did not receive fatal error: %s", tc.err)
			}
		})
	}
}

//func TestTo(t *testing.T) {
//	for _, tc := range []struct {
//		name string
//		to   int
//		in   string
//		want string
//	}{
//		{
//			name: "Basic",
//			to:   3,
//			in: `0,1
//1,2
//2,3
//2,4
//`,
//			want: `	1 -> 2
//	2 -> 3
//`,
//		},
//		{
//			name: "CyclesBasic",
//			to:   2,
//			in: `0,1
//1,2
//2,1
//`,
//			want: `
//	1 -> 2`,
//		},
//		{
//			name: "CyclesWithSplit",
//			to:   3,
//			in: `0,1
//1,2
//2,1
//2,3
//2,4
//`,
//			want: `	1 -> 2
//	2 -> 3
//	2 -> 1
//`,
//		},
//		{
//			name: "TwoSplits",
//			to:   8,
//			in: `0,1
//1,2
//1,3
//2,4
//3,4
//4,5
//4,6
//6,7
//5,7
//7,8
//`,
//			want: `	1 -> 2
//	2 -> 4
//	4 -> 5
//	5 -> 7
//	7 -> 8
//	1 -> 3
//	3 -> 4
//	4 -> 6
//	6 -> 7
//`,
//		},
//		{
//			name: "TwoSplitsMidpoint",
//			to:   4,
//			in: `0,1
//1,2
//1,3
//2,4
//3,4
//4,5
//4,6
//6,7
//5,7
//7,8
//`,
//			want: `	1 -> 2
//	2 -> 4
//	1 -> 3
//	3 -> 4
//`,
//		},
//		{
//			name: "TwoSplitsLeaf",
//			to:   5,
//			in: `0,1
//1,2
//1,3
//2,4
//3,4
//4,5
//4,6
//6,7
//5,7
//7,8
//`,
//			want: `	1 -> 2
//	2 -> 4
//	4 -> 5
//	1 -> 3
//	3 -> 4
//`,
//		},
//		{
//			name: "TwoSplitsGap",
//			to:   10,
//			in: `0,1
//1,2
//1,3
//2,4
//3,4
//4,5
//5,6
//5,7
//6,8
//7,9
//8,10
//9,10
//`,
//			want: `	1 -> 2
//	2 -> 4
//	4 -> 5
//	5 -> 6
//	6 -> 8
//	8 -> 10
//	1 -> 3
//	3 -> 4
//	5 -> 7
//	7 -> 9
//	9 -> 10
//`,
//		},
//	} {
//		t.Run(tc.name, func(t *testing.T) {
//			db, err := mockQuery([]string{"fromcomid", "tocomid"}, query, tc.in)
//			if err != nil {
//				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
//			}
//			defer db.Close()
//
//			out := bytes.Buffer{}
//			catchments, err := fromDB(db, tq)
//			if err != nil {
//				t.Fatalf("unexpected error %s while creating graph", err)
//			}
//
//			subCatchments, err := catchments.To(tc.to)
//			if err != nil {
//				t.Fatalf("unexpected error while building sub graph: %s", err)
//			}
//
//			if err := subCatchments.PrintTo(&out, tc.to); err != nil {
//				t.Fatalf("unexpected error while printing graph %s", err)
//			}
//
//			got := sortByNewLine(out.String())
//			want := sortByNewLine(tc.want)
//			if got != want {
//				t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
//			}
//		})
//	}
//}

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
