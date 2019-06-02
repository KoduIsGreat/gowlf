package main

import (
	"database/sql"
	"fmt"
	"io"
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
	from *PathTraversed
	seen map[int]bool
}

type Needle struct{
	id int
	paths []*PathTraversed
}

func newVertex(id int) *Vertex {
	return &Vertex{id: id, edges: map[int]*Vertex{}}
}

func (g *Graph) print(out io.Writer) error {
	return g.printDfs(out, map[int]bool{}, g.root)
}

func (p* PathTraversed) isNeighbor(needle int) bool{
	for _, edge := range p.edges {
		fmt.Printf("is node %d a neighbor of %d? edge: %d\n",p.id, needle, edge.id)
		if edge.id == needle {
			return true
		}
	}
	return false
}
func (g *Graph) printDfs(out io.Writer, visited map[int]bool, cursor *Vertex) error {
	if visited[cursor.id] {
		return nil // stop
	}
	visited[cursor.id] = true
	for _, edge := range cursor.edges {
		if g.vertices[edge.id] {
			if cursor.id != 0 {
				if _, err := fmt.Fprintf(out, "\t%d -> %d\n", cursor.id, edge.id); err != nil {
					return err
				}
			}
			if err := g.printDfs(out, visited, edge); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Graph) To(needle int) *Graph {
	newVertices := map[int]bool{}
	need := &Needle{id: needle, paths: []*PathTraversed{}}
	g.toDfs(map[int]bool{}, &PathTraversed{Vertex: g.root},need)
	for _, path := range need.paths {
		for path.from != nil {
			newVertices[path.id] = true
			path = path.from
		}
	}
	return &Graph{root: g.root, vertices: newVertices}
}

func (g *Graph) toDfs(visited map[int]bool,cursor *PathTraversed, needle *Needle) {
	if visited[cursor.id]{
		return
	}
	visited[cursor.id] = true
	if cursor.id == needle.id{
		needle.paths = append(needle.paths, cursor)
	} else {
		for _, edge := range cursor.edges {
			g.toDfs(visited,&PathTraversed{Vertex: edge, from: cursor},needle)
		}
	}
	visited[cursor.id] = false
	return
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
		return nil, fmt.Errorf("error with query: %s", err)
	}
	for rows.Next() {
		if err := rows.Scan(&from, &to); err != nil {
			return nil, fmt.Errorf("error reading row: %s",err)
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
		return nil, fmt.Errorf("root %d does not exist", rootId)
	}

	return &Graph{vertices: vertices, root: rootVertex}, nil
}
