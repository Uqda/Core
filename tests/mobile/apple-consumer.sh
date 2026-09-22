#!/bin/sh
set -eu

XCFRAMEWORK=${1:-Uqda.xcframework}
test -d "$XCFRAMEWORK"

FRAMEWORK=$(find "$XCFRAMEWORK" -type d -name Uqda.framework -path '*macos*' | head -n 1)
test -n "$FRAMEWORK"
test -s "$FRAMEWORK/Headers/Uqda.h"
grep -q '@interface MobileUqda' "$FRAMEWORK/Headers/Uqda.h"

STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT HUP INT TERM
cat > "$STAGE/consumer.m" <<'EOF'
@import Uqda;

int main(void) {
    MobileUqda *node = [[MobileUqda alloc] init];
    return node == nil;
}
EOF

xcrun clang -fmodules -fobjc-arc -framework Foundation \
  -F "$(dirname "$FRAMEWORK")" -framework Uqda \
  -c "$STAGE/consumer.m" -o "$STAGE/consumer.o"
test -s "$STAGE/consumer.o"
