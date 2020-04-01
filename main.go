package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbPath = flag.String("db", "./db/una.sqlite", "Path to database e.g : --db ./path/to/my/db.sqlite")
var query = "SELECT distinct fromcomid, tocomid FROM catchment_navigation INNER JOIN catchments ON catchments.comid = catchment_navigation.fromcomid or catchments.comid = catchment_navigation.tocomid;"

func main() {

	flag.Usage = func() {
		log.Println("gwlf [--db] | dot -Tpng -o outfile.png")
	}
	flag.Parse()
	var out bytes.Buffer
	if *dbPath == "" {
		log.Fatal(fmt.Errorf("--db must be set"))
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	catchmentNetwork, err := fromDB(db, query)
	if err != nil {
		log.Fatal(err)
	}

	catchmentNetwork.sprint(&out)

	if _, err := out.WriteTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
