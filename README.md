# gh-pr-count

gh extension to count the number of PRs with the same label as the directory name

## Motivation

Suppose you are a monorepo user and want to know the statistics of which services are actively developed.
You assume that each Pull Request _has a label with the name of the directory_ that contains the files changed by the Pull Request. ([actions/labeler](https://github.com/actions/labeler) can help you to add the label automatically.)

In that case, this gh-extension displays the number of labels that match the directory (service name).

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
gh pr-count 2023-10-01 2023-10-31 # Count the number of PRs in October 2023
gh pr-count 2023-11-01            # Count the number of PRs since November 1st, 2023 until now
```

Output example:

```sh
backend,54
docs,20
frontend,10
```

Please note that the order of the output is not guaranteed. You can sort the output by sort command to get the consistent result.

## Environment Variables

| Name               | Description                                                                                                                            |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| `GH_REPO`          | The repository to query. Defaults to the current repository.                                                                           |
| `SEARCH_QUERY`     | The search query to use. If you exclude a PR with `dependencies` label and by bot, set `"-label:dependencies -author:app/my-cool-bot"` |
| `MAX_CONCURRENTCY` | The maximum number of concurrentcy. Defaults to 50.                                                                                    |

## Known limitations

#### Only the first 100 search results are available

`gh list` command returns only the first 100 results. Set the environment variable `SEARCH_QUERY` or change the 'since' argument to return less than 100 results.
