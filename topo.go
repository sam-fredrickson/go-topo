package topo

import "errors"

// ErrCyclicDependency is returned when the graph contains a cycle.
var ErrCyclicDependency = errors.New("cyclic dependency detected")

// node represents any element that can have dependencies on other elements.
//
// If node A has deps [X, Y, Z], it means A depends on X, Y, and Z,
// so X, Y, and Z must be processed before A.
type node[T comparable] struct {
	value T
	deps  []T
}

// Graph represents a collection of nodes with their dependencies.
type Graph[T comparable] struct {
	nodes []node[T]
}

// AddNode adds a node to the graph with its dependencies.
func (g *Graph[T]) AddNode(value T, deps []T) {
	g.nodes = append(g.nodes, node[T]{
		value: value,
		deps:  deps,
	})
}

// SortByLayers performs a topological sort of the graph, returning layers
// where each layer contains nodes that can be processed in parallel.
// Each layer must be processed before the next layer.
func (g *Graph[T]) SortByLayers() ([][]T, error) {
	// node values to dependencies
	dependsOn := make(map[T][]T)
	// reverse: node values to nodes that depend on them
	dependedOnBy := make(map[T][]T)
	// all values in the graph
	allValues := make(map[T]bool)
	for _, node := range g.nodes {
		allValues[node.value] = true
		dependsOn[node.value] = node.deps
		for _, dep := range node.deps {
			allValues[dep] = true
			dependedOnBy[dep] = append(dependedOnBy[dep], node.value)
		}
	}

	// find nodes with no dependencies;
	// these form the first layer
	var currentLayer []T
	for value := range allValues {
		deps, exists := dependsOn[value]
		if !exists || len(deps) == 0 {
			currentLayer = append(currentLayer, value)
		}
	}

	// process the graph layer by layer
	var result [][]T
	visited := make(map[T]bool)
	for len(currentLayer) > 0 {
		// invariant: current layer is finalized
		result = append(result, currentLayer)

		// mark these nodes as visited
		for _, value := range currentLayer {
			visited[value] = true
		}

		// find the next layer - nodes that depend on the current layer nodes
		// and have all their dependencies resolved
		var nextLayer []T
		layerDepMap := make(map[T]bool) // To avoid duplicates in the next layer

		for _, value := range currentLayer {
			// find nodes that depend on this one
			for _, dependent := range dependedOnBy[value] {
				// skip if already processed or in next layer
				if visited[dependent] || layerDepMap[dependent] {
					continue
				}

				// check if all dependencies of this dependent node are visited
				allDepsVisited := true
				for _, dep := range dependsOn[dependent] {
					if !visited[dep] {
						allDepsVisited = false
						break
					}
				}

				// if all dependencies are satisfied, add to next layer
				if allDepsVisited {
					nextLayer = append(nextLayer, dependent)
					layerDepMap[dependent] = true
				}
			}
		}

		currentLayer = nextLayer
	}

	// check if all nodes were visited
	for value := range allValues {
		if !visited[value] {
			// if this is a node in our original graph (not just a dependency)
			if _, exists := dependsOn[value]; exists && len(dependsOn[value]) > 0 {
				return nil, ErrCyclicDependency
			}
		}
	}

	return result, nil
}
