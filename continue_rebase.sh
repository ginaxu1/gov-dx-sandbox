#!/bin/bash
cd /Users/tmp/opendif-mvp

# Ensure all files are staged
git add -A

# Check for unmerged files
UNMERGED=$(git ls-files -u | wc -l | tr -d ' ')
if [ "$UNMERGED" -gt 0 ]; then
    echo "ERROR: There are $UNMERGED unmerged files"
    git ls-files -u
    exit 1
fi

# Continue the rebase
echo "Continuing rebase..."
git rebase --continue

# Check if rebase completed
if [ ! -d .git/rebase-merge ] && [ ! -d .git/rebase-apply ]; then
    echo "SUCCESS: Rebase completed!"
    git log --oneline -5
else
    echo "Rebase still in progress"
    cat .git/rebase-merge/msgnum 2>/dev/null || cat .git/rebase-apply/next 2>/dev/null
fi

