#!/bin/sh

echo "🔍 Auditing test coverage per package..."
echo ""

go list -buildvcs=false ./... | grep -v /vendor/ | while read pkg; do
    if grep -q "$pkg" ./tests/coverage/coverage-unit.out; then
        echo "✅ Covered: $pkg"
    else
        echo "❌ Missing from coverage: $pkg"
    fi
done
