package graph

import (
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := NewGraph()

	if g == nil {
		t.Fatal("NewGraph returned nil")
	}

	if g.nodes == nil {
		t.Error("Nodes map not initialized")
	}

	if g.edges == nil {
		t.Error("Edges slice not initialized")
	}
}

func TestAddNode(t *testing.T) {
	g := NewGraph()

	node := g.AddNode("node1", "service1", "/api/test", "GET", "test data")

	if node == nil {
		t.Fatal("AddNode returned nil")
	}

	if node.ID != "node1" {
		t.Errorf("Expected ID 'node1', got '%s'", node.ID)
	}

	if node.Service != "service1" {
		t.Errorf("Expected service 'service1', got '%s'", node.Service)
	}

	if g.NodeCount() != 1 {
		t.Errorf("Expected node count 1, got %d", g.NodeCount())
	}
}

func TestAddNodeDuplicate(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node1", "service2", "/api/other", "POST", nil)

	if node1 != node2 {
		t.Error("Duplicate node ID should return existing node")
	}

	if g.NodeCount() != 1 {
		t.Errorf("Expected node count 1, got %d", g.NodeCount())
	}
}

func TestAddEdge(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)

	edge := g.AddEdge(node1, node2, 1.0, nil)

	if edge == nil {
		t.Fatal("AddEdge returned nil")
	}

	if edge.From != node1 {
		t.Error("Edge from node not set correctly")
	}

	if edge.To != node2 {
		t.Error("Edge to node not set correctly")
	}

	if edge.Weight != 1.0 {
		t.Errorf("Expected weight 1.0, got %f", edge.Weight)
	}

	if g.EdgeCount() != 1 {
		t.Errorf("Expected edge count 1, got %d", g.EdgeCount())
	}
}

func TestGetNode(t *testing.T) {
	g := NewGraph()

	g.AddNode("node1", "service1", "/api/test", "GET", nil)

	node, exists := g.GetNode("node1")
	if !exists {
		t.Error("Node not found")
	}

	if node.ID != "node1" {
		t.Errorf("Expected ID 'node1', got '%s'", node.ID)
	}

	_, exists = g.GetNode("nonexistent")
	if exists {
		t.Error("Nonexistent node should not be found")
	}
}

func TestGetOutgoingEdges(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)
	node3 := g.AddNode("node3", "service3", "/api/third", "GET", nil)

	g.AddEdge(node1, node2, 1.0, nil)
	g.AddEdge(node1, node3, 1.5, nil)

	edges := g.GetOutgoingEdges(node1)

	if len(edges) != 2 {
		t.Errorf("Expected 2 outgoing edges, got %d", len(edges))
	}
}

func TestGetIncomingEdges(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)
	node3 := g.AddNode("node3", "service3", "/api/third", "GET", nil)

	g.AddEdge(node1, node3, 1.0, nil)
	g.AddEdge(node2, node3, 1.0, nil)

	edges := g.GetIncomingEdges(node3)

	if len(edges) != 2 {
		t.Errorf("Expected 2 incoming edges, got %d", len(edges))
	}
}

func TestHasCycleNoCycle(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)
	node3 := g.AddNode("node3", "service3", "/api/third", "GET", nil)

	g.AddEdge(node1, node2, 1.0, nil)
	g.AddEdge(node2, node3, 1.0, nil)

	if g.HasCycle() {
		t.Error("Graph should not have a cycle")
	}
}

func TestHasCycleWithCycle(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)
	node3 := g.AddNode("node3", "service3", "/api/third", "GET", nil)

	g.AddEdge(node1, node2, 1.0, nil)
	g.AddEdge(node2, node3, 1.0, nil)
	g.AddEdge(node3, node1, 1.0, nil) // Creates cycle

	if !g.HasCycle() {
		t.Error("Graph should have a cycle")
	}
}

func TestTopologicalSort(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)
	node3 := g.AddNode("node3", "service3", "/api/third", "GET", nil)

	g.AddEdge(node1, node2, 1.0, nil)
	g.AddEdge(node2, node3, 1.0, nil)

	sorted, err := g.TopologicalSort()

	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(sorted))
	}

	// Check that dependencies come before dependents
	nodeMap := make(map[string]int)
	for i, node := range sorted {
		nodeMap[node.ID] = i
	}

	if nodeMap["node1"] >= nodeMap["node2"] {
		t.Error("node1 should come before node2")
	}

	if nodeMap["node2"] >= nodeMap["node3"] {
		t.Error("node2 should come before node3")
	}
}

func TestTopologicalSortWithCycle(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)

	g.AddEdge(node1, node2, 1.0, nil)
	g.AddEdge(node2, node1, 1.0, nil)

	_, err := g.TopologicalSort()

	if err == nil {
		t.Error("Expected error for cyclic graph")
	}
}

func TestFindAllPaths(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)
	node3 := g.AddNode("node3", "service3", "/api/third", "GET", nil)
	node4 := g.AddNode("node4", "service4", "/api/fourth", "GET", nil)

	// Create two paths from node1 to node4
	g.AddEdge(node1, node2, 1.0, nil)
	g.AddEdge(node2, node4, 1.0, nil)
	g.AddEdge(node1, node3, 1.0, nil)
	g.AddEdge(node3, node4, 1.0, nil)

	paths := g.FindAllPaths("node1", "node4", 5)

	if len(paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(paths))
	}
}

func TestGetAllNodes(t *testing.T) {
	g := NewGraph()

	g.AddNode("node1", "service1", "/api/test", "GET", nil)
	g.AddNode("node2", "service2", "/api/other", "GET", nil)
	g.AddNode("node3", "service3", "/api/third", "GET", nil)

	nodes := g.GetAllNodes()

	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
}

func TestGetAllEdges(t *testing.T) {
	g := NewGraph()

	node1 := g.AddNode("node1", "service1", "/api/test", "GET", nil)
	node2 := g.AddNode("node2", "service2", "/api/other", "GET", nil)

	g.AddEdge(node1, node2, 1.0, nil)
	g.AddEdge(node2, node1, 2.0, nil)

	edges := g.GetAllEdges()

	if len(edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(edges))
	}
}
