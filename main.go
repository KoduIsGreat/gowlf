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
var query = flag.String("q","SELECT fromcomid, tocomid FROM catchment_navigation;","a SQL query that returns an adjacency list")
var allpaths = flag.Int("ap", -1, "display all paths to the given node")
func main() {

	flag.Usage = func() {
		log.Println("gwlf [--db] [--q] [--ap] | dot -Tpng -o outfile.png")
	}
	flag.Parse()
	if err := gwlf(*dbPath, *query, *allpaths); err != nil {
		log.Println(err.Error())
	}
}

func gwlf(dbPath, query string, ap int) error{
	var out bytes.Buffer
	if dbPath == "" {
		return fmt.Errorf("--db must be set")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	catchmentNetwork, err := toFromDb(db, query)
	if err != nil {
		return err
	}
	if ap > 0 {
		catchmentNetwork = catchmentNetwork.subNetwork(ap)
	}
	catchmentNetwork.dotprint(&out)

	if _, err := out.WriteTo(os.Stdout); err != nil {
		return err
	}
	return nil
}
