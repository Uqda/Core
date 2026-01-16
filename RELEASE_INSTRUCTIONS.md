# Release Instructions for v0.1.0

## Prerequisites

- Git repository initialized and connected to https://github.com/Uqda/Core.git
- All changes committed
- Go 1.22+ installed (for building)

## Step 1: Verify Everything is Ready

```bash
# Check git status
git status

# Verify all files are committed
git log --oneline -5

# Test the build
./build

# Verify binaries
./uqda -version
./uqdactl -help
```

## Step 2: Create and Push the Tag

```bash
# Create annotated tag for v0.1.0
git tag -a v0.1.0 -m "Release v0.1.0 - Initial Uqda Core release

First release of Uqda Core, forked from Yggdrasil Network.
Rebranded and independently maintained by the Uqda team.

Release date: January 16, 2026"

# Push the tag to GitHub
git push origin v0.1.0

# Or push all tags
git push origin --tags
```

## Step 3: Create GitHub Release

### Option A: Using GitHub Web Interface

1. Go to https://github.com/Uqda/Core/releases
2. Click "Draft a new release"
3. Select tag: `v0.1.0`
4. Release title: `v0.1.0 - Initial Release`
5. Description: Copy from `RELEASE_v0.1.0.md`
6. Check "Set as the latest release"
7. Click "Publish release"

### Option B: Using GitHub CLI (gh)

```bash
# Install gh if not already installed
# Then create release:
gh release create v0.1.0 \
  --title "v0.1.0 - Initial Release" \
  --notes-file RELEASE_v0.1.0.md \
  --latest
```

## Step 4: Verify Release

1. Check that the release appears at: https://github.com/Uqda/Core/releases
2. Verify that GitHub Actions workflows run (if configured)
3. Test downloading and installing the release

## Step 5: Update Main Branch (Optional)

If you want to update the main branch with release information:

```bash
# Update version in README or other files if needed
# Commit and push
git add .
git commit -m "Update for v0.1.0 release"
git push origin main
```

## Complete Command Sequence

```bash
# 1. Verify everything
git status
./build

# 2. Create and push tag
git tag -a v0.1.0 -m "Release v0.1.0 - Initial Uqda Core release"
git push origin v0.1.0

# 3. Create release (using GitHub web interface or gh CLI)
# Visit: https://github.com/Uqda/Core/releases/new
```

## Troubleshooting

### Tag Already Exists

If the tag already exists locally:

```bash
# Delete local tag
git tag -d v0.1.0

# Delete remote tag (if pushed)
git push origin --delete v0.1.0

# Then recreate
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0
```

### Build Fails

Make sure:
- Go 1.22+ is installed
- All dependencies are available
- Run `go mod tidy` if needed

### Tag Not Showing on GitHub

- Wait a few moments for GitHub to process
- Check that you pushed to the correct remote: `git remote -v`
- Verify tag exists: `git tag -l`

---

**Release Date**: January 16, 2026  
**Version**: v0.1.0

