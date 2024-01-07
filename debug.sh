#!/bin/bash
since="2023-12-01"
GH_BASE="develop"
GH_LABEL="api"

gh pr list \
  --base "${GH_BASE}" \
  --repo "${GH_REPO}" \
  --label ${GH_LABEL} \
  --search "merged:>=${since} -label:dependencies -author:app/quipper-monorepo-ci" \
  --limit 100 \
  --json author \
  --template '{{range .}}{{tablerow .author.login }}{{end}}'