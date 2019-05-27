package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
)

type Graph struct {
	root     *Vertex
	vertices map[int]bool
}

type Vertex struct {
	id    int
	edges map[int]*Vertex
}

type PathTraversed struct {
	*Vertex
	from *Vertex
}

func newVertex(id int) *Vertex {
	return &Vertex{id: id, edges: map[int]*Vertex{}}
}

func (g *Graph) print(out io.Writer) error {
	return g.printDfs(out, map[int]bool{}, g.root)
}

func (g *Graph) printDfs(out io.Writer, visited map[int]bool, cursor *Vertex) error {
	if visited[cursor.id] {
		return nil // stop
	}
	visited[cursor.id] = true
	for _, edge := range cursor.edges {
		if cursor.id != 0 {
			if _, err := fmt.Fprintf(out, "\t%d -> %d\n", cursor.id, edge.id); err != nil {
				return err
			}
		}
		if err := g.printDfs(out, visited, edge); err != nil {
			return err
		}
	}
	return nil
}

func newGraph(db *sql.DB) (*Graph, error) {
	return newNetwork(db, 0)
}

func newNetwork(db *sql.DB, rootId int) (*Graph, error) {
	vertexMap := map[int]*Vertex{}
	var (
		from int
		to   int
	)

	rows, err := db.Query("SELECT distinct fromcomid, tocomid FROM catchment_navigation INNER JOIN catchments ON catchments.comid = catchment_navigation.fromcomid or catchments.comid = catchment_navigation.tocomid;")
	if err != nil {
		log.Fatal(err)
		return nil, fmt.Errorf("Error with query")
	}
	for rows.Next() {
		if err := rows.Scan(&from, &to); err != nil {
			log.Fatal(err)
			return nil, fmt.Errorf("Error reading row")
		}

		v, ok := vertexMap[from]
		if !ok {
			v = newVertex(from)
			vertexMap[v.id] = v
		}

		u, ok := vertexMap[to]
		if !ok {
			u = newVertex(to)
			vertexMap[u.id] = u
		}
		v.edges[u.id] = u

	}
	vertices := map[int]bool{}
	for k := range vertexMap {
		vertices[k] = true
	}

	rootVertex, ok := vertexMap[rootId]
	if !ok {
		return nil, fmt.Errorf("Catchment %s does not exist", rootId)
	}

	return &Graph{vertices: vertices, root: rootVertex}, nil
}
