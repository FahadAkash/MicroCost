package graph

import (
	"fmt"

	"github.com/microcost/microcost/pkg/models"
)

// Graph represents a directed graph structure
type Graph struct {
	nodes map[string]*Node
	edges []*Edge
}

// Node represents a vertex in the graph
type Node struct {
	ID       string
	Service  string
	Endpoint string
	Method   string
	Data     interface{}
}

// Edge represents a directed edge in the graph
type Edge struct {
	From   *Node
	To     *Node
	Weight float64
	Data   *models.Dependency
}

// NewGraph creates a new empty graph
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		edges: make([]*Edge, 0),
	}
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(id, service, endpoint, method string, data interface{}) *Node {
	if node, exists := g.nodes[id]; exists {
		return node
	}

	node := &Node{
		ID:       id,
		Service:  service,
		Endpoint: endpoint,
		Method:   method,
		Data:     data,
	}
	g.nodes[id] = node
	return node
}

// AddEdge adds an edge between two nodes
func (g *Graph) AddEdge(from, to *Node, weight float64, dep *models.Dependency) *Edge {
	edge := &Edge{
		From:   from,
		To:     to,
		Weight: weight,
		Data:   dep,
	}
	g.edges = append(g.edges, edge)
	return edge
}

// GetNode retrieves a node by ID
func (g *Graph) GetNode(id string) (*Node, bool) {
	node, exists := g.nodes[id]
	return node, exists
}

// GetOutgoingEdges returns all edges originating from a node
func (g *Graph) GetOutgoingEdges(node *Node) []*Edge {
	edges := make([]*Edge, 0)
	for _, edge := range g.edges {
		if edge.From.ID == node.ID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// GetIncomingEdges returns all edges pointing to a node
func (g *Graph) GetIncomingEdges(node *Node) []*Edge {
	edges := make([]*Edge, 0)
	for _, edge := range g.edges {
		if edge.To.ID == node.ID {
			edges = append(edges, edge)
		}
	}
	return edges
}

// HasCycle detects if the graph contains a cycle using DFS
func (g *Graph) HasCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for id := range g.nodes {
		if !visited[id] {
			if g.hasCycleDFS(id, visited, recStack) {
				return true
			}
		}
	}
	return false
}

func (g *Graph) hasCycleDFS(nodeID string, visited, recStack map[string]bool) bool {
	visited[nodeID] = true
	recStack[nodeID] = true

	node, _ := g.GetNode(nodeID)
	for _, edge := range g.GetOutgoingEdges(node) {
		if !visited[edge.To.ID] {
			if g.hasCycleDFS(edge.To.ID, visited, recStack) {
				return true
			}
		} else if recStack[edge.To.ID] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}

// TopologicalSort performs topological sorting of the graph
// Returns nodes in dependency order (dependencies come before dependents)
func (g *Graph) TopologicalSort() ([]*Node, error) {
	if g.HasCycle() {
		return nil, fmt.Errorf("graph contains cycles, cannot perform topological sort")
	}

	inDegree := make(map[string]int)
	for id := range g.nodes {
		inDegree[id] = 0
	}

	for _, edge := range g.edges {
		inDegree[edge.To.ID]++
	}

	queue := make([]*Node, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			node, _ := g.GetNode(id)
			queue = append(queue, node)
		}
	}

	sorted := make([]*Node, 0)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, node)

		for _, edge := range g.GetOutgoingEdges(node) {
			inDegree[edge.To.ID]--
			if inDegree[edge.To.ID] == 0 {
				queue = append(queue, edge.To)
			}
		}
	}

	return sorted, nil
}

// FindAllPaths finds all paths from start to end node
func (g *Graph) FindAllPaths(startID, endID string, maxDepth int) [][]*Node {
	start, exists := g.GetNode(startID)
	if !exists {
		return nil
	}

	paths := make([][]*Node, 0)
	currentPath := make([]*Node, 0)
	visited := make(map[string]bool)

	g.findPathsDFS(start, endID, maxDepth, currentPath, visited, &paths)
	return paths
}

func (g *Graph) findPathsDFS(current *Node, endID string, maxDepth int, currentPath []*Node, visited map[string]bool, paths *[][]*Node) {
	if len(currentPath) > maxDepth {
		return
	}

	currentPath = append(currentPath, current)
	visited[current.ID] = true

	if current.ID == endID {
		pathCopy := make([]*Node, len(currentPath))
		copy(pathCopy, currentPath)
		*paths = append(*paths, pathCopy)
	} else {
		for _, edge := range g.GetOutgoingEdges(current) {
			if !visited[edge.To.ID] {
				g.findPathsDFS(edge.To, endID, maxDepth, currentPath, visited, paths)
			}
		}
	}

	visited[current.ID] = false
}

// GetAllNodes returns all nodes in the graph
func (g *Graph) GetAllNodes() []*Node {
	nodes := make([]*Node, 0, len(g.nodes))
	for _, node := range g.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetAllEdges returns all edges in the graph
func (g *Graph) GetAllEdges() []*Edge {
	return g.edges
}

// NodeCount returns the number of nodes in the graph
func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

// EdgeCount returns the number of edges in the graph
func (g *Graph) EdgeCount() int {
	return len(g.edges)
}
