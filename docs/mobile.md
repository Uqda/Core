# Mobile bindings

The exported binding type is `mobile.Uqda`; its methods include `StartJSON`,
`StartAutoconfigure`, `Send`, `Recv` and `Stop`. For persistent identity, store the
configuration returned by `GenerateConfigJSON` securely and reuse it:

```go
var node mobile.Uqda
if err := node.StartJSON(savedConfig); err != nil {
    return err
}
defer node.Stop()
```

Generated Java and Objective-C/Swift bindings expose the Uqda type name.
`GetVersion()` returns the machine-readable release version.

From the root, `./contrib/mobile/build -a` builds `uqda.aar` using gomobile and
the Android SDK/NDK. On macOS, `./contrib/mobile/build -i` builds
`Uqda.xcframework` for iOS/macOS using Xcode and gomobile. Install the Go mobile
binding tools before invoking the script. Record their versions for reproducible
builds; the script's Go mobile dependency resolution can update module files.

`go test ./contrib/mobile` exercises Go behavior. Release validation also builds
the AAR/XCFramework and compiles minimal Java and Objective-C consumers through
`tests/mobile/android-consumer.sh` and `tests/mobile/apple-consumer.sh`.
Generated bindings are not committed.
