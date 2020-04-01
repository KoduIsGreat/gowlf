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
func (n network) sprint(out *bytes.Buffer) {
	var sb strings.Builder
	for node, edges := range n {
		var sb2 strings.Builder
		sb2.WriteString(fmt.Sprintf("%d", node))
		for edge := range edges {
			sb2.WriteString(fmt.Sprintf(" %d", edge))
		}
		sb2.WriteString("\n")
		s := sb2.String()
		sb.WriteString(s)
	}
	out.WriteString(sb.String())
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
	t := n.transpose()
	visit := func(node int) {
		if _, ok := seen[node]; !ok {
			for edge := range t[node] {
				sub.addEdges(node, edge)
				q = append(q, edge)
			}
			seen[node] = struct{}{}
		}
	}
	for len(q) > 0 {
		node := q[0]
		q = q[1:]
		visit(node)
	}
	return sub.transpose()
}

// creates a network from a *sql.DB provided a query
func fromDB(db *sql.DB, q string) (network, error) {
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
		network.addEdges(from, to)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}
	return network, nil
}
