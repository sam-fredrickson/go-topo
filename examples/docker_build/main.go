package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/sam-fredrickson/go-topo"
)

// ImageMetadata represents the metadata for a Docker image.
type ImageMetadata struct {
	Name         string   `json:"name"`
	Path         string   `json:"path"`
	Dependencies []string `json:"dependencies"`
}

// RepositoryMetadata represents the metadata for all images in the repository.
type RepositoryMetadata struct {
	Images []ImageMetadata `json:"images"`
}

//go:embed metadata.json
var metadataBytes []byte

func main() {
	// in real life you might read from a file, but in this example we
	// embed it to make running the program simpler.
	// metadataBytes, err := os.ReadFile("metadata.json")
	// if err != nil {
	// 	fmt.Printf("Error reading metadata file: %v\n", err)
	// 	os.Exit(1)
	// }

	var repoMetadata RepositoryMetadata
	if err := json.Unmarshal(metadataBytes, &repoMetadata); err != nil {
		fmt.Printf("Error parsing metadata: %v\n", err)
		os.Exit(1)
	}

	imagesByName := make(map[string]ImageMetadata)
	for _, img := range repoMetadata.Images {
		imagesByName[img.Name] = img
	}

	var g topo.Graph[string]
	for _, img := range repoMetadata.Images {
		g.AddNode(img.Name, img.Dependencies)
	}

	layers, err := g.SortByLayers()
	if err != nil {
		fmt.Printf("Error sorting dependencies: %v\n", err)
		os.Exit(1)
	}

	// build each layer in sequence, with parallel builds within each layer
	for i, layer := range layers {
		fmt.Printf("\n--- Building Layer %d ---\n", i+1)

		var wg sync.WaitGroup
		errChan := make(chan error, len(layer))

		for _, imageName := range layer {
			wg.Add(1)
			go func() {
				defer wg.Done()
				img, exists := imagesByName[imageName]
				if !exists {
					errChan <- fmt.Errorf("image %s metadata not found", imageName)
					return
				}

				fmt.Printf("Building image: %s\n", imageName)

				err := buildImage(img.Path, imageName)
				if err != nil {
					errChan <- fmt.Errorf("error building image %s: %v", imageName, err)
					return
				}

				fmt.Printf("Successfully built image: %s\n", imageName)
			}()
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		fmt.Printf("--- Layer %d completed ---\n", i+1)
	}

	fmt.Println("\nAll images built successfully!")
}

func buildImage(path, name string) error {
	dockerfilePath := filepath.Join(path, "Dockerfile")
	cmd := exec.Command("docker", "build", "-t", name, "-f", dockerfilePath, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// return cmd.Run()
	return nil
}
