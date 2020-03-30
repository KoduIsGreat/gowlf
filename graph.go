package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
)

// represents a unique set of catchments or comIds
type catchSet map[int]struct{}

// Represents a catchment network which is a directed acyclic graph of watershed basins
type network map[int]catchSet

// Prints the graph in the form of textual words the first word is the ancestor node and any
// words proceeding on the same line are its descendants
func(n network) sprint(out *bytes.Buffer) {
	for node, edges := range n {
		var sb strings.Builder
		for edge := range edges {
			sb.WriteString(fmt.Sprintf("%d\t", edge))
		}
		s := fmt.Sprintf("%d\t%s\n", node, sb.String())
		out.WriteString(s)
	}
}

// Adds a node to the network
func (n network) addNode(node int) catchSet {
	edges := n[node]
	if edges == nil {
		edges = make(catchSet)
		n[node] =edges
	}
	return edges
}

// Adds an edge to the network, if the from node doesn't exist in the network, it is added
func (n network) addEdges(from int, tos ...int)  {
	edges := n.addNode(from)
	for _, to := range tos {
		n.addNode(to)
		edges[to] = struct{}{}
	}
}

func (n network) transpose() network {
	rev := make(network)
	for node, edges := range n {
		rev.addNode(node)
		for succ := range edges {
			rev.addEdges(succ, node)
		}
	}
	return rev
}


// creates a network from a *sql.DB provided a query
func fromDB(db *sql.DB, q string) (network, error) {
	network := make(network)
	var (
		from int
		to   int
	)
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("error with query: %w", err)
	}
	for rows.Next() {
		if err := rows.Scan(&from, &to); err != nil {
			return nil, fmt.Errorf("error reading row: %w", err)
		}
		network.addEdges(from, to)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return network, nil
}

