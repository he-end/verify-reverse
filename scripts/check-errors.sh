#!/usr/bin/env bash
set -euo pipefail

echo "=== Checking for anti-patterns ==="

echo ""
echo "--- stdlib log usage (should not appear outside test files) ---"
grep -rn '"log"' auth/ --include='*.go' | grep -v '_test.go' | grep -v '/log/' | grep -v '/goruntime/' || true

echo ""
echo "--- log.Println / log.Printf / log.Fatalf (should be replaced) ---"
grep -rn 'log\.\(Print\|Fatal\)' auth/ --include='*.go' | grep -v '_test.go' || true

echo ""
echo "--- fmt.Println in non-test code (should use logger) ---"
grep -rn 'fmt\.Println' auth/ --include='*.go' | grep -v '_test.go' | grep -v '/log/' || true

echo ""
echo "--- bare return err in service/repository (potential missing context) ---"
grep -rn 'return err$' auth/service/ auth/repository/ --include='*.go' | grep -v '_test.go' || true

echo ""
echo "--- bare return nil, err in service/repository ---"
grep -rn 'return nil, err$' auth/service/ auth/repository/ --include='*.go' | grep -v '_test.go' || true

echo ""
echo "=== Check complete ==="
