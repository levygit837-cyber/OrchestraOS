#!/usr/bin/env bash

# pre-push.sh — install to .git/hooks/pre-push to block direct pushes to main.
# Usage: cp scripts/git/pre-push.sh .git/hooks/pre-push && chmod +x .git/hooks/pre-push
#
# This hook is called with the following arguments:
#   $1 Name of the remote to which the push is being done
#   $2 URL to which the push is being done
#
# If pushing without a named remote, arguments will be empty but stdin
# receives lines of the form: <local ref> <local sha1> <remote ref> <remote sha1>

while read local_ref local_sha remote_ref remote_sha; do
    if [ "$remote_ref" = "refs/heads/main" ]; then
        echo ""
        echo "❌ PUSH BLOCKED: direct push to 'main' is forbidden."
        echo ""
        echo "Required workflow:"
        echo "   1. git checkout -b feature/your-change"
        echo "   2. git commit -m 'your changes'"
        echo "   3. git push origin feature/your-change"
        echo "   4. Open a Pull Request on GitHub"
        echo "   5. Wait for CI to pass before merging"
        echo ""
        exit 1
    fi
done
