# Branch Protection Without GitHub Paid Features

Since GitHub Rulesets and (in some cases) Classic Branch Protection require a paid Team/Org plan, we use **client-side git hooks** and **controlled scripts** to enforce the same rules locally.

These measures work entirely on your machine and do NOT depend on GitHub settings.

---

## Defense Layer 1: Pre-Push Hook (Blocks Direct Push to Main)

The `pre-push` hook runs **before every `git push`** and aborts if you try to push `main` directly.

### Install

```bash
cp scripts/git/pre-push.sh .git/hooks/pre-push
chmod +x .git/hooks/pre-push
```

### What it does

If you run `git push origin main`, the hook outputs:

```
❌ PUSH BLOCKED: direct push to 'main' is forbidden.

Required workflow:
   1. git checkout -b feature/your-change
   2. git commit -m 'your changes'
   3. git push origin feature/your-change
   4. Open a Pull Request on GitHub
   5. Wait for CI to pass before merging
```

And the push is cancelled.

---

## Defense Layer 2: Pre-Commit Hook (Validates Code Before Commit)

Runs `go vet`, architecture tests, and contract verification before allowing any commit.

### Install

```bash
cp scripts/git/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## Defense Layer 3: Safe Commit Script (Prevents Commits on Main)

`scripts/git/safe-commit.sh` is a wrapper that:
1. Detects if you are on `main`
2. Automatically creates a feature branch if so
3. Runs all validations
4. Commits only after everything passes

### Usage

```bash
./scripts/git/safe-commit.sh "feat: add billing module"
```

---

## Defense Layer 4: CI as Auditor

Even if someone bypasses all local hooks (e.g., by using `git push --no-verify` or editing files directly on GitHub), the CI still runs on every push. The `.github/workflows/ci.yml` will:

- Run tests
- Run architecture boundary checks
- Run contract verification
- Run linter

If the code is bad, CI fails. Without branch protection, CI failure is a **red flag** during review, not a hard block. For solo projects, this is usually sufficient because you (or the LLM) will see the failure and fix it before considering the work done.

---

## One-Line Setup (Run Once After Clone)

```bash
# Install all hooks and tools
cp scripts/git/pre-commit.sh .git/hooks/pre-commit
cp scripts/git/pre-push.sh .git/hooks/pre-push
chmod +x .git/hooks/pre-commit .git/hooks/pre-push
```

---

## For LLM Agents

Add to your LLM system prompt or `AGENTS.md`:

> **NEVER commit or push directly to `main`.**
> 
> Before committing, run `./scripts/git/safe-commit.sh "message"`. This script will automatically create a feature branch if you are on `main`, run all validations, and commit safely.
> 
> After committing, push the feature branch and open a Pull Request. Wait for CI to pass.
