# Branch Protection Setup

To ensure that tests must pass before merging to master, follow these steps:

## Configure Branch Protection Rules

1. Go to your repository on GitHub
2. Click on **Settings** → **Branches**
3. Under "Branch protection rules", click **Add rule**
4. Configure the following settings:

### Branch name pattern
```
master
```
or
```
main
```

### Required settings

#### Protect matching branches
- ✅ **Require a pull request before merging**
  - ✅ Require approvals (optional, recommended: 1)
  - ✅ Dismiss stale pull request approvals when new commits are pushed

- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - Select the following status checks:
    - `Run Tests (1.21)`
    - `Run Tests (1.22)`
    - `Run Tests (1.23)`
    - `Run Linters`

- ✅ **Require conversation resolution before merging** (optional)

- ✅ **Do not allow bypassing the above settings** (recommended)

## Result

After configuring these rules:
- No one can push directly to master
- All changes must go through pull requests
- Tests must pass before merging
- Failed tests will block the merge

## Testing the Configuration

1. Create a new branch: `git checkout -b test-branch`
2. Make some changes
3. Push the branch: `git push origin test-branch`
4. Create a Pull Request to master
5. GitHub Actions will automatically run tests
6. Merge button will be disabled until all tests pass
