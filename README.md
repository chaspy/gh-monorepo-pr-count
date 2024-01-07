# gh-monorepo-pr-count

gh extension to count the number of PRs with the same label as the directory name for monorepo

## Motivation

Suppose you are a monorepo user and want to know the statistics of which services are actively developed.
You assume that each Pull Request _has a label with the name of the directory_ that contains the files changed by the Pull Request.

`gh-monorepo-pr-count` will count a number of PR by label for each directory in your monorepo.

### How to add a label to a PR

#### [Danger](https://github.com/danger/danger)

Recommended.

```ruby
labels = (git.added_files.to_a + git.deleted_files.to_a + git.modified_files.to_a).map {|f| Pathname.new(f).descend.first.to_s }.uniq

repo = 'user/monorepo' # TODO: Update the repository name

if labels.count > 0
  number = github.pr_json['number']
  current_labels = github.api.labels_for_issue(repo, number).map(&:name)
  missing_labels = labels - current_labels
  unless missing_labels.empty?
    begin
      github.api.add_labels_to_an_issue(repo, number, missing_labels)
    rescue Octokit::Error => e
      message <<~MESSAGE
        Failed to add a label to the PR #{number}
        #{e.class}: #{e.message}
      MESSAGE
    end
  end
end
```

#### [actions/labeler](https://github.com/actions/labeler)

You should update your `.github/labeler.yml` each time a new directory is added

```yaml
docs:
  - changed-files:
      - any-glob-to-any-file: docs/**
backend:
  - changed-files:
      - any-glob-to-any-file: backend/**
```

## Installation

```sh
gh extension install chaspy/gh-monorepo-pr-count
```

To upgrade,

```sh
gh extension upgrade chaspy/gh-monorepo-pr-count
```

## Usage

```sh
Usage of ./gh-monorepo-pr-count:
  -debug-url
    	Optional: Print debug url
  -since string
    	Required: Search PRs merged since this date. Format: yyyy-mm-dd
  -uniq-author
    	Optional: Count a number of PR for each directory by uniq author
  -until string
    	Optional: Search PRs merged until this date. Format: yyyy-mm-dd (default is today)
```

Output example:

```sh
backend,54
docs,20
frontend,10
```

Please note that the order of the output is not guaranteed. You can sort the output by sort command to get the consistent result.

## Environment Variables

| Name              | Description                                                                                                                            |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `GH_REPO`         | The repository to query. Defaults to the current repository.                                                                           |
| `SEARCH_QUERY`    | The search query to use. If you exclude a PR with `dependencies` label and by bot, set `"-label:dependencies -author:app/my-cool-bot"` |
| `MAX_CONCURRENCY` | The maximum number of concurrentcy. Defaults to 50.                                                                                    |

## Known limitations

#### Only the first 100 search results are available

`gh list` command returns only the first 100 results. Set the environment variable `SEARCH_QUERY` or change the 'since' argument to return less than 100 results.
