# gh-pr-count

gh extension to count the number of PRs with the same label as the directory name

## Environment Variables

| Name           | Description                                                  |
| -------------- | ------------------------------------------------------------ |
| `GH_REPO`      | The repository to query. Defaults to the current repository. |
| `SEARCH_QUERY` | The search query to use.                                     |

## Known limitations

#### Only the first 100 search results are available

`gh list` command returns only the first 100 results. Set the environment variable `SEARCH_QUERY` or change the 'since' argument to return less than 100 results.
