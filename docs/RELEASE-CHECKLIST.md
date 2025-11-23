# Release Checklist

Use this checklist before creating a new release.

## Pre-Release Verification

### Build & Compilation
- [ ] `make build` completes without errors
- [ ] `make test` passes all tests
- [ ] No compiler warnings

### Scenario Tests
- [ ] `.scratch/test_all_visualization.sh` passes
- [ ] `.scratch/test_integration.sh` passes
- [ ] All existing scenario tests pass:
  - [ ] `test_update_delete_contact.sh`
  - [ ] `test_company_crud.sh`
  - [ ] `test_graphs.sh`

### Manual Testing
- [ ] TUI launches and navigates correctly (`pagen`)
  - [ ] Tab switching works
  - [ ] Arrow key navigation works
  - [ ] Enter shows details
  - [ ] Edit view works
  - [ ] Graph view works
  - [ ] Delete confirmations work
- [ ] Terminal dashboard displays (`pagen viz`)
  - [ ] Stats are correct
  - [ ] Pipeline bars render
  - [ ] Attention items show when applicable
- [ ] Web UI serves correctly (`pagen web`)
  - [ ] Dashboard loads at http://localhost:8080
  - [ ] All navigation links work
  - [ ] HTMX partials load without page refresh
  - [ ] Search filters work
  - [ ] Graph generation works
- [ ] GraphViz graphs generate (`pagen viz graph`)
  - [ ] Contact graphs work
  - [ ] Company graphs work
  - [ ] Pipeline graphs work
- [ ] MCP server starts (`pagen mcp`)
  - [ ] Server responds to requests
  - [ ] All 20 tools are available

### Documentation
- [ ] README.md is up to date
  - [ ] Command examples are accurate
  - [ ] MCP tool count is correct (20)
  - [ ] New features are documented
- [ ] All plan documents are complete
- [ ] CLAUDE.md is accurate (if applicable)

### Code Quality
- [ ] No TODO comments in production code
- [ ] All functions have ABOUTME comments
- [ ] Error handling is consistent
- [ ] No debug logging left in code

### Dependencies
- [ ] `go.mod` and `go.sum` are clean
- [ ] All dependencies are necessary
- [ ] No version conflicts

## Release Process

### Version Bump
- [ ] Update version in relevant files
- [ ] Update CHANGELOG.md (if exists)

### Git
- [ ] All changes committed
- [ ] Working directory is clean
- [ ] On main branch (or release branch)

### Tag & Push
- [ ] Create git tag: `git tag -a v0.X.0 -m "Release v0.X.0"`
- [ ] Push tag: `git push origin v0.X.0`
- [ ] Push commits: `git push`

### Build Release Binaries
- [ ] Build for Linux: `GOOS=linux GOARCH=amd64 go build -o pagen-linux-amd64`
- [ ] Build for macOS: `GOOS=darwin GOARCH=amd64 go build -o pagen-darwin-amd64`
- [ ] Build for macOS ARM: `GOOS=darwin GOARCH=arm64 go build -o pagen-darwin-arm64`
- [ ] Test each binary on target platform

### GitHub Release (if applicable)
- [ ] Create GitHub release from tag
- [ ] Upload binaries
- [ ] Write release notes highlighting new features

## Post-Release
- [ ] Verify release is downloadable
- [ ] Test installation from release
- [ ] Update project documentation links
- [ ] Announce release (if applicable)

## Rollback Plan
If issues are discovered:
1. Document the issue
2. Create hotfix branch
3. Fix and test
4. Create patch release (v0.X.1)
5. Mark broken release as pre-release/draft
