# Branching

Uqda Core keeps exactly two long-lived branches.

| Branch | Purpose | Who changes it |
|---|---|---|
| `main` | The stable branch. Releases are tagged from it and published images are built from it. | Only by merging a tested `beta` into it. |
| `beta` | Where all new work lands: features, fixes, dependency updates. | Pull requests from short-lived work branches. |

## Everyday work

1. Branch from `beta` (for example `fix/handshake-limits`).
2. Open a pull request against `beta`. Dependabot is configured to do the same.
3. Merge when CI is green. The work branch is deleted after merge, so only
   `main` and `beta` remain.

## Promoting beta to main

Promote only when `beta` has a green CI run on its latest commit and the change
has been tried on real nodes where it matters (peering, TUN, installers).

```sh
git fetch origin
git checkout main
git merge --ff-only origin/beta
git push origin main
```

Use a fast-forward so `main` and `beta` point at the same commits. If the
fast-forward fails, something was committed to `main` directly: bring it into
`beta` first, test, then promote. Tag releases on `main` after promotion; see
[releasing](releasing.md).
