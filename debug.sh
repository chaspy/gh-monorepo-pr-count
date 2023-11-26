#!/bin/bash
since="2023-10-01"
gh pr list --base "${GH_BASE}" --repo "${GH_REPO}" --label ${GH_LABEL} --search "merged:>=${since} -label:dependencies" --limit 100
