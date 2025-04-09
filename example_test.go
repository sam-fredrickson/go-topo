package topo_test

import (
	"cmp"
	"errors"
	"slices"
	"testing"

	"github.com/sam-fredrickson/go-topo"
)

// TestExamples runs tests of example use cases.
func TestExamples(t *testing.T) {
	t.Run("ContainerDependencies", func(t *testing.T) {
		runExample(t, example[string]{
			init: func(g *topo.Graph[string]) {
				g.AddNode("base-image", []string{})
				g.AddNode("app-image", []string{"base-image"})
				g.AddNode("cache-image", []string{"base-image"})
				g.AddNode("test-image", []string{"app-image", "cache-image"})
				g.AddNode("dev-image", []string{"base-image"})
			},
			expected: [][]string{
				{"base-image"},
				{"app-image", "cache-image", "dev-image"},
				{"test-image"},
			},
			validate: func(t *testing.T, layers [][]string) {
				validateLayerDependencies(t, layers, map[string][]string{
					"app-image":   {"base-image"},
					"cache-image": {"base-image"},
					"test-image":  {"app-image", "cache-image"},
					"dev-image":   {"base-image"},
				})
			},
		})
	})

	t.Run("ComplexDependencies", func(t *testing.T) {
		runExample(t, example[string]{
			init: func(g *topo.Graph[string]) {
				g.AddNode("database", []string{})
				g.AddNode("redis", []string{})
				g.AddNode("api", []string{"database", "redis"})
				g.AddNode("auth-service", []string{"database"})
				g.AddNode("background-worker", []string{"database", "redis"})
				g.AddNode("frontend", []string{"api", "auth-service"})
				g.AddNode("monitoring", []string{"api", "background-worker", "frontend"})
			},
			expected: [][]string{
				{"database", "redis"},
				{"api", "auth-service", "background-worker"},
				{"frontend"},
				{"monitoring"},
			},
			validate: func(t *testing.T, layers [][]string) {
				validateLayerDependencies(t, layers, map[string][]string{
					"api":               {"database", "redis"},
					"auth-service":      {"database"},
					"background-worker": {"database", "redis"},
					"frontend":          {"api", "auth-service"},
					"monitoring":        {"api", "background-worker", "frontend"},
				})
			},
		})
	})

	t.Run("CyclicDependency", func(t *testing.T) {
		runExample(t, example[string]{
			init: func(g *topo.Graph[string]) {
				// Add nodes with a cyclic dependency
				g.AddNode("service-a", []string{"service-b"})
				g.AddNode("service-b", []string{"service-c"})
				g.AddNode("service-c", []string{"service-a"})
			},
			expectErr:   true,
			expectedErr: topo.ErrCyclicDependency,
		})
	})

	t.Run("GenericTypes", func(t *testing.T) {
		runExample(t, example[int]{
			init: func(g *topo.Graph[int]) {
				g.AddNode(1, []int{})
				g.AddNode(2, []int{1})
				g.AddNode(3, []int{1})
				g.AddNode(4, []int{2, 3})
			},
			expected: [][]int{
				{1},
				{2, 3},
				{4},
			},
		})
	})
}

type example[T cmp.Ordered] struct {
	init        func(graph *topo.Graph[T])
	expectErr   bool
	expectedErr error
	expected    [][]T
	validate    func(t *testing.T, layers [][]T)
}

func runExample[T cmp.Ordered](t *testing.T, e example[T]) {
	var g topo.Graph[T]
	if e.init != nil {
		e.init(&g)
	}

	layers, err := g.SortByLayers()
	if e.expectErr {
		// verify that we got the expected error
		if err == nil {
			t.Errorf("Expected error (%v), got nil", e.expectedErr)
		} else if !errors.Is(err, e.expectedErr) {
			t.Errorf("Expected error (%v), got: %v", e.expectedErr, err)
		}
	} else if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(layers) != len(e.expected) {
		t.Errorf("Expected %d layers, got %d",
			len(e.expected), len(layers))
	}

	// sort the layers for deterministic comparison
	for i := range layers {
		slices.Sort(layers[i])
	}

	// validate layer contents
	for n := range layers {
		if !slices.Equal(layers[n], e.expected[n]) {
			t.Errorf("Expected layer %d to be %v, got %v", n,
				e.expected, layers[0])
		}
	}

	if e.validate != nil {
		e.validate(t, layers)
	}
}

// validateLayerDependencies ensures that all dependencies are processed before dependent nodes.
func validateLayerDependencies[T comparable](
	t *testing.T, layers [][]T, dependencies map[T][]T,
) {
	// track which layer each node is in
	nodeLayer := make(map[T]int)
	for i, layer := range layers {
		for _, node := range layer {
			nodeLayer[node] = i
		}
	}

	// check that all dependencies are in earlier layers
	for node, deps := range dependencies {
		nodeLayerIndex, exists := nodeLayer[node]
		if !exists {
			t.Errorf("node %v not found in any layer", node)
			continue
		}
		for _, dep := range deps {
			depLayerIndex, exists := nodeLayer[dep]
			if !exists {
				t.Errorf("Dependency %v of node %v not found in any layer", dep, node)
				continue
			}

			if depLayerIndex >= nodeLayerIndex {
				t.Errorf("Dependency violation: %v (layer %d) depends on %v (layer %d)",
					node, nodeLayerIndex+1, dep, depLayerIndex+1)
			}
		}
	}
}
