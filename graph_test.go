package main

import (
	"bytes"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPrint(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	query := "SELECT distinct fromcomid, tocomid FROM catchment_navigation INNER JOIN catchments ON catchments.comid = catchment_navigation.fromcomid or catchments.comid = catchment_navigation.tocomid;"
	columns := []string{"fromcomid", "tocomid"}
	in := `0,307562200
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
`
	want := `	307578700 -> 307586600
	307586600 -> 307586700
	307586700 -> 307586800
	307586800 -> 307592300
	307601400 -> 307586800
	307562200 -> 307586600
`
	out := bytes.Buffer{}
	mock.ExpectQuery(query).WillReturnRows(mock.NewRows(columns).FromCSVString(in))
	catchments, err := newGraph(db)
	if err != nil {
		t.Fatalf("unexpected error %s while creating graph", err)
	}
	if err := catchments.print(&out); err != nil {
		t.Fatalf("unexpected error %s while printing graph", err)
	}

	got := out.String()
	if got != want {
		t.Fatalf("Test failed: \n got %s \n\n want %s", got, want)
	}

}
