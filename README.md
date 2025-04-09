# go-topo

A generic Go library for layered topological sorting of dependencies.

## Overview

go-topo provides a way to sort elements with dependencies into layers, where:

1. Elements in each layer can be processed in parallel
2. All elements in a layer must be processed before moving to the next layer
3. Dependencies are always processed before the elements that depend on them

This is particularly useful for build systems, deployment orchestration, task scheduling,
and any other scenario where you need to process items with dependencies efficiently.

## Installation

```bash
go get github.com/sam-fredrickson/go-topo
```

## Features

- Generic implementation that works with any comparable type
  - Strings, integers, pointers, etc.
- Efficient layered topological sorting algorithm
- Cycle detection
- Simple, clean API

## Use Cases

- Docker image build pipeline with dependent images
- Deployment orchestration where services have dependencies
- Task scheduling with prerequisite tasks
- Processing stages in a data pipeline
- Resolving package dependencies

## Usage

Basic usage example:

```go
package main

import (
	"fmt"
	"github.com/sam-fredrickson/go-topo"
)

func main() {
	// Create a new dependency graph
	var g topo.Graph[string]

	// Add nodes with their dependencies
	g.AddNode("base-image", []string{})
	g.AddNode("app-image", []string{"base-image"})
	g.AddNode("cache-image", []string{"base-image"})
	g.AddNode("test-image", []string{"app-image", "cache-image"})

	// Perform the layered topological sort
	layers, err := g.SortByLayers()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Process the layers
	for i, layer := range layers {
		fmt.Printf("Layer %d: %v\n", i+1, layer)

		// In a real application, you might process each layer in parallel
		// For example:
		// var wg sync.WaitGroup
		// for _, node := range layer {
		//     wg.Add(1)
		//     go func() {
		//         defer wg.Done()
		//         processNode(node)
		//     }()
		// }
		// wg.Wait()
	}
}
```
