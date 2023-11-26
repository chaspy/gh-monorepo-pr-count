package main

import (
	"fmt"
	"log"

	"os"
	"path/filepath"

	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

func main() {
	// Get directory name under the current directory

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != "." && !filepath.HasPrefix(path, ".") {
			fmt.Println(path)
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	// Get the current repository
	repo, err := repository.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current repository: %s/%s\n", repo.Owner, repo.Name)

	targetRepo := repo.Owner + "/" + repo.Name
	fmt.Printf("Target repository: %s\n", targetRepo)
	// To overwrite the target repository, use the GH_REPO environment variable

	// Count a number of PRs
	fmt.Println("counting a number of PRs...")

	// Test
	repos := []string{"api,", "docs", "bin"}
	for _, repo := range repos {
		fmt.Println(repo)

		prList, _, err := gh.Exec("pr", "list", "--repo", targetRepo, "--label", repo)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(prList.String())
	}
}
