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

	// Count a number of PRs
	fmt.Println("counting a number of PRs...")

	prList, _, err := gh.Exec("pr", "list", "--repo", "cli/cli", "--limit", "5")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(prList.String())
}
