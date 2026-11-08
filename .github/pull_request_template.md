<!-- Open pull requests against `beta`. `main` only receives tested `beta` merges. -->

## What and why

<!-- The user-visible problem, and what changes. -->

## How it was tested

<!-- Commands you ran and their results. Say which tests you wrote and which you actually executed. -->

- [ ] `go test -race ./...`
- [ ] `go vet ./...` and `gofmt`
- [ ] Docs updated (README, `docs/`, CHANGELOG) if behaviour changed

## Compatibility

- [ ] No change to the Yggdrasil wire format, or the change is optional and negotiated
- [ ] Existing configurations and node identities keep working
