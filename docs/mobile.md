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

The type rename from `Yggdrasil` is an API break. Rebuild generated Java and
Objective-C/Swift bindings, update client type names and remove cached old binding
artifacts. No legacy alias is exported. `GetVersion()` exposes the machine version.

From the root, `./contrib/mobile/build -a` builds `uqda.aar` using gomobile and
the Android SDK/NDK. On macOS, `./contrib/mobile/build -i` builds
`Uqda.xcframework` for iOS/macOS using Xcode and gomobile. Install the Go mobile
binding tools before invoking the script. Record their versions for reproducible
builds; the script's Go mobile dependency resolution can update module files.

`go test ./contrib/mobile` exercises Go behavior, but does not validate generated
binding consumers. Release validation must compile both AAR/framework artifacts
and consuming applications with the new API. Generated bindings are not committed.
