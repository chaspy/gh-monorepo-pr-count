# gh-pr-count

gh extension to count the number of PRs with the same label as the directory name

## Installation

```sh
gh extension install chaspy/gh-pr-count
```

To upgrade,

```sh
gh extension upgrade chaspy/gh-pr-count
```

## Usage

```sh
gh pr-count YYYY-MM-DD
```

## Environment Variables

| Name               | Description                                                                                                                            |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| `GH_REPO`          | The repository to query. Defaults to the current repository.                                                                           |
| `SEARCH_QUERY`     | The search query to use. If you exclude a PR with `dependencies` label and by bot, set `"-label:dependencies -author:app/my-cool-bot"` |
| `MAX_CONCURRENTCY` | The maximum number of concurrentcy. Defaults to 50.                                                                                    |

## Known limitations

#### Only the first 100 search results are available

`gh list` command returns only the first 100 results. Set the environment variable `SEARCH_QUERY` or change the 'since' argument to return less than 100 results.
