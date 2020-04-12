package main

import (
	"bytes"
	"database/sql"
	"fmt"
)

// represents a unique set of catchments or comIds
type catchSet map[int]struct{}

// Represents a catchment network which is a directed acyclic graph of watershed basins
type network map[int]catchSet

// Prints the graph in the form of textual words the first word is the ancestor node and any
// words proceeding on the same line are its descendants
func (n network) print(out *bytes.Buffer) {
	for node, edges := range n {
		out.WriteString(fmt.Sprintf("%d", node))
		for edge := range edges {
			out.WriteString(fmt.Sprintf(" %d", edge))
		}
		out.WriteString("\n")
	}
}

// Prints the graph in Graphviz dot notation: https://www.graphviz.org/doc/info/lang.html
func(n network) dotprint(out *bytes.Buffer) {
	out.WriteString(fmt.Sprint("digraph {\n"))
	for node, edges := range n {
		for edge := range edges{
			e := fmt.Sprintf("\t%d -> %d\n", node, edge)
			out.WriteString(e)
		}
	}
	out.WriteString("}")
}

// Adds a node to the network
func (n network) addNode(node int) catchSet {
	edges := n[node]
	if edges == nil {
		edges = make(catchSet)
		n[node] = edges
	}
	return edges
}

// Adds an edge to the network, if the from node doesn't exist in the network, it is added
func (n network) addEdges(from int, tos ...int) {
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

// produces a subnetwork given an input node
func (n network) subNetwork(node int) network {
	sub := make(network)
	seen := make(catchSet)
	q := make([]int, 0)
	q = append(q, node)
	visit := func(node int) {
		if _, ok := seen[node]; !ok {
			for edge := range n[node] {
				sub.addEdges(node, edge)
				q = append(q, edge)
			}
			seen[node] = struct{}{}
		}
	}
	// while we have items in our queue continue to visit nodes
	for len(q) > 0 {
		node := q[0]
		q = q[1:]
		visit(node)
	}
	return sub.transpose()
}

// creates a network from a *sql.DB provided a query
func toFromDb(db *sql.DB, q string) (network, error) {
	network := make(network)
	var from, to int
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("error with query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.Scan(&from, &to); err != nil {
			return nil, fmt.Errorf("error reading row: %w", err)
		}
		network.addEdges(to, from)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return network, nil
}

