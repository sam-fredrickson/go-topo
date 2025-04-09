package topo_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/sam-fredrickson/go-topo"
)

// TestSortByLayers runs some basic test cases.
func TestSortByLayers(t *testing.T) {
	tests := []struct {
		name          string
		nodes         map[string][]string
		expected      [][]string
		expectErr     bool
		expectedError error
	}{
		{
			name: "simple linear dependencies",
			nodes: map[string][]string{
				"A": {},
				"B": {"A"},
				"C": {"B"},
			},
			expected:  [][]string{{"A"}, {"B"}, {"C"}},
			expectErr: false,
		},
		{
			name: "parallel dependencies",
			nodes: map[string][]string{
				"A": {},
				"B": {},
				"C": {"A", "B"},
				"D": {"A", "B"},
			},
			expected:  [][]string{{"A", "B"}, {"C", "D"}},
			expectErr: false,
		},
		{
			name: "complex dependencies",
			nodes: map[string][]string{
				"A": {},
				"B": {},
				"C": {"A"},
				"D": {"B"},
				"E": {"C", "D"},
				"F": {"A"},
			},
			expected:  [][]string{{"A", "B"}, {"C", "D", "F"}, {"E"}},
			expectErr: false,
		},
		{
			name: "cyclic dependencies",
			nodes: map[string][]string{
				"A": {"C"},
				"B": {"A"},
				"C": {"B"},
			},
			expectErr:     true,
			expectedError: topo.ErrCyclicDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var g topo.Graph[string]
			for node, deps := range tt.nodes {
				g.AddNode(node, deps)
			}

			result, err := g.SortByLayers()

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error %v, got nil", tt.expectedError)
					return
				}
				if !errors.Is(err, tt.expectedError) {
					t.Errorf("Expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// sort each layer to ensure deterministic comparison, since
			// the order within a layer doesn't matter
			for i := range result {
				sort.Strings(result[i])
			}

			expectedSorted := make([][]string, len(tt.expected))
			for i, layer := range tt.expected {
				expectedSorted[i] = make([]string, len(layer))
				copy(expectedSorted[i], layer)
				sort.Strings(expectedSorted[i])
			}

			if !reflect.DeepEqual(result, expectedSorted) {
				t.Errorf("Expected %v, got %v", expectedSorted, result)
			}
		})
	}
}
