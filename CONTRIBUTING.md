# Contributing

Install Go 1.25 or newer, clone `Uqda/Core`, and run `./build` (`build.bat` on
Windows). Read the [architecture](docs/architecture.md), [compatibility contract](docs/compatibility.md)
and [testing instructions](docs/testing.md).

Use `gofmt`, add regression coverage for behavioral fixes, and run build, vet,
tests and vulnerability scanning. Preserve wire compatibility, identity and
upstream attribution. Describe the user-visible problem, behavior change and
actual validation in pull requests; distinguish authored tests from executed tests.
Report security issues privately as described in [SECURITY](SECURITY.md).

Use Uqda for the product and عُقدة in Arabic prose; Yggdrasil refers to the
network, protocol or upstream source. Commands and paths remain `uqda`,
`uqdactl`, `uqda.conf` and `Uqda/Core`. Release version rules are in
[releasing](docs/releasing.md); platform details are in [packaging](docs/packaging.md).
