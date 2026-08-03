#!/usr/bin/env bash
set -e

echo "🍬 Compilando p2t CLI para múltiplos ambientes..."
mkdir -p dist

VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-X github.com/Wather17/p2t/cmd.Version=${VERSION} -X github.com/Wather17/p2t/cmd.Commit=${COMMIT} -X github.com/Wather17/p2t/cmd.Date=${DATE} -s -w"

echo " └─ Compilando para Linux (amd64)..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/p2t-linux-amd64 ./cmd/p2t

echo " └─ Compilando para Linux (arm64)..."
GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o dist/p2t-linux-arm64 ./cmd/p2t

echo " └─ Compilando para Windows (amd64)..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o dist/p2t-windows-amd64.exe ./cmd/p2t

echo "✅ Compilação concluída com sucesso! Binários gerados na pasta dist/"
