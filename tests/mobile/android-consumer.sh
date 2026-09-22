#!/bin/sh
set -eu

AAR=${1:-uqda.aar}
test -s "$AAR"

STAGE=$(mktemp -d)
trap 'rm -rf "$STAGE"' EXIT HUP INT TERM
unzip -q "$AAR" -d "$STAGE/aar"
test -s "$STAGE/aar/classes.jar"

UQDA_CLASS=$(jar tf "$STAGE/aar/classes.jar" | awk '/\/Uqda.class$/ { print; exit }')
test -n "$UQDA_CLASS"
PACKAGE_PATH=${UQDA_CLASS%/Uqda.class}
PACKAGE=$(printf '%s' "$PACKAGE_PATH" | tr / .)

mkdir -p "$STAGE/src" "$STAGE/classes"
cat > "$STAGE/src/Consumer.java" <<EOF
import ${PACKAGE}.Uqda;

public final class Consumer {
    private final Uqda node = new Uqda();

    public Uqda node() {
        return node;
    }
}
EOF

javac -encoding UTF-8 -source 8 -target 8 \
  -cp "$STAGE/aar/classes.jar" \
  -d "$STAGE/classes" "$STAGE/src/Consumer.java"
test -s "$STAGE/classes/Consumer.class"
