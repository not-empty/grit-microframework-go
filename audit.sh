#!/bin/sh
reset

COVERAGE_DIR="./tests/coverage"
COVERAGE_OUT="$COVERAGE_DIR/coverage-unit.out"
COVERAGE_TXT="$COVERAGE_DIR/coverage-unit.txt"
COVERAGE_HTML="$COVERAGE_DIR/coverage-unit.html"

echo "✅ Running unit tests with coverage..."
echo ""

# Step 1: Build package list and exclude unwanted ones
PACKAGES=$(go list -buildvcs=false ./... | grep -vE '/vendor/|tests/|/cmd$|/app/context$|/app/repository/models$|/app/router/domains$|/app/router/routes')
COVERPKG=$(echo "$PACKAGES" | paste -sd, -)

# Step 2: Run tests
go test -v -tags=unit -coverprofile="$COVERAGE_OUT" -coverpkg="$COVERPKG" ./tests/...

# Step 3: Generate reports
mkdir -p "$COVERAGE_DIR"
go tool cover -func="$COVERAGE_OUT" -o "$COVERAGE_TXT"
go tool cover -html="$COVERAGE_OUT" -o "$COVERAGE_HTML"
sed -i 's|<title>handler: Go Coverage Report</title>|<title>Unit Go Coverage Report</title>|g' "$COVERAGE_HTML"

# Step 4: Audit per-package coverage
echo ""
echo "🔍 Auditing Unit test coverage per package..."
echo ""

if [ ! -f "$COVERAGE_TXT" ]; then
    echo "❌ Coverage report not found: $COVERAGE_TXT"
    exit 1
fi

for pkg in $PACKAGES; do
    pattern=$(echo "$pkg" | sed 's/\//\\\//g')

    percent=$(grep "^$pattern" "$COVERAGE_TXT" | awk '{sum+=$3; count++} END {if (count>0) printf "%.1f", sum/count; else print "0.0"}')

    emoji="❌"
    if [ "$percent" = "100.0" ]; then
        emoji="✅"
    elif [ "$(echo "$percent > 0" | bc)" -eq 1 ]; then
        emoji="🔸"
    fi

    printf "%s %s - %s%%\n" "$emoji" "$pkg" "$percent"
done

# Step 5: Total coverage
total_line=$(tail -n 1 "$COVERAGE_TXT" | grep -E '^total:')
if [ -n "$total_line" ]; then
    total_percent=$(echo "$total_line" | awk '{print $3}')
    echo ""
    echo "📊 Total project coverage: $total_percent"
fi

echo ""
echo "🧪 View detailed HTML coverage report:"
echo "👉  ./tests/coverage/coverage-unit.html"
