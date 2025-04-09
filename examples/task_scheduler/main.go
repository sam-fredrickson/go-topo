package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/sam-fredrickson/go-topo"
)

// Task represents a job with dependencies that needs to be executed.
type Task struct {
	ID          string
	Description string
	Duration    time.Duration
}

func main() {
	fmt.Println("Task Scheduler with Layered Topological Sort")
	fmt.Println("===========================================")

	tasks := map[string]Task{
		"setup-db":      {ID: "setup-db", Description: "Initialize database schema", Duration: 2 * time.Second},
		"load-data":     {ID: "load-data", Description: "Load initial data", Duration: 3 * time.Second},
		"api-server":    {ID: "api-server", Description: "Start API server", Duration: 1 * time.Second},
		"worker":        {ID: "worker", Description: "Start background worker", Duration: 1 * time.Second},
		"cache":         {ID: "cache", Description: "Initialize cache", Duration: 1 * time.Second},
		"notifications": {ID: "notifications", Description: "Setup notification service", Duration: 2 * time.Second},
		"frontend":      {ID: "frontend", Description: "Start frontend server", Duration: 1 * time.Second},
		"monitoring":    {ID: "monitoring", Description: "Start monitoring service", Duration: 1 * time.Second},
		"load-balancer": {ID: "load-balancer", Description: "Configure load balancer", Duration: 2 * time.Second},
		"final-checks":  {ID: "final-checks", Description: "Run system checks", Duration: 1 * time.Second},
	}

	var g topo.Graph[string]

	g.AddNode("setup-db", []string{})
	g.AddNode("load-data", []string{"setup-db"})
	g.AddNode("api-server", []string{"load-data"})
	g.AddNode("worker", []string{"load-data"})
	g.AddNode("cache", []string{"setup-db"})
	g.AddNode("notifications", []string{"worker"})
	g.AddNode("frontend", []string{"api-server", "cache"})
	g.AddNode("monitoring", []string{"api-server", "worker", "cache"})
	g.AddNode("load-balancer", []string{"api-server", "frontend"})
	g.AddNode("final-checks", []string{"frontend", "monitoring", "load-balancer", "notifications"})

	// calculate the execution layers
	layers, err := g.SortByLayers()
	if err != nil {
		fmt.Printf("Error sorting dependencies: %v\n", err)
		return
	}

	fmt.Println("\nTask Execution Plan:")
	for i, layer := range layers {
		fmt.Printf("Layer %d: %v\n", i+1, layer)
	}

	// execute the tasks layer by layer
	fmt.Println("\nExecuting tasks:")
	startTime := time.Now()

	for i, layer := range layers {
		fmt.Printf("\n--- Layer %d ---\n", i+1)
		layerStart := time.Now()

		// execute tasks in this layer concurrently
		var wg sync.WaitGroup
		for _, taskID := range layer {
			wg.Add(1)
			go func() {
				defer wg.Done()
				task := tasks[taskID]
				fmt.Printf("Starting task: %s - %s\n", taskID, task.Description)

				// Simulate task execution
				time.Sleep(task.Duration)

				fmt.Printf("Completed task: %s (took %v)\n", taskID, task.Duration)
			}()
		}

		// wait for all tasks in this layer to complete
		wg.Wait()

		layerDuration := time.Since(layerStart)
		fmt.Printf("--- Layer %d completed in %v ---\n", i+1, layerDuration)
	}

	totalDuration := time.Since(startTime)
	fmt.Printf("\nAll tasks completed in %v\n", totalDuration)

	// calculate the theoretical sequential execution time
	var sequentialDuration time.Duration
	for _, task := range tasks {
		sequentialDuration += task.Duration
	}

	fmt.Printf("Sequential execution would take: %v\n", sequentialDuration)
	fmt.Printf("Parallel execution saved: %v\n", sequentialDuration-totalDuration)
}
