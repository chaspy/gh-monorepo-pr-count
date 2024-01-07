set -eu
go build .
mv gh-monorepo-pr-count "../../${GH_REPO}"