package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var dbPath = flag.String("db","", "Path to database e.g : --db ./path/to/my/db.sqlite")
func main() {

	flag.Usage = func() {
		log.Println("gwlf [--db] | dot -Tpng -o outfile.png")
	}
	flag.Parse()
	var out bytes.Buffer
	if *dbPath =="" {
		log.Fatal(fmt.Errorf("--db must be set"))
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	catchmentNetwork, err := newGraph(db)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := out.Write([]byte("digraph gomodgraph {\n")); err != nil {
		log.Fatal(err)
	}

	if err:= catchmentNetwork.print(&out); err !=nil{
		log.Fatal(err)
	}

	if _, err := out.Write([]byte("}\n")); err != nil {
		log.Fatal(err)
	}

	if _, err := out.WriteTo(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

