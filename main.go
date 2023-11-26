package main

import (
	"fmt"
	"log"
	gh "github.com/cli/go-gh/v2"
	"github.com/cli/go-gh/v2/pkg/repository"
)

func main() {
	// Get the current repository
	repo, err := repository.Current()
	fmt.Printf("Current repository: %s/%s\n", repo.Owner, repo.Name)

	targetRepo := repo.Owner + "/" + repo.Name
	fmt.Printf("Target repository: %s\n", targetRepo)

	// Count a number of PRs
	fmt.Println("counting a number of PRs...")

	prList, _, err := gh.Exec("pr", "list", "--repo", targetRepo, "--limit", "5")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(prList.String())
}
