#!/usr/bin/env bash
set -e

INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="p2t"

echo "🍬 Instalando p2t CLI..."

mkdir -p "${INSTALL_DIR}"

if [ -f "dist/p2t-linux-amd64" ]; then
    echo " └─ Copiando binário pré-compilado de dist/..."
    cp dist/p2t-linux-amd64 "${INSTALL_DIR}/${BINARY_NAME}"
else
    echo " └─ Compilando executável..."
    go build -o "${INSTALL_DIR}/${BINARY_NAME}" ./cmd/p2t
fi

chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo "✅ p2t instalado com sucesso em: ${INSTALL_DIR}/${BINARY_NAME}"
echo ""
echo "Certifique-se de que '${INSTALL_DIR}' está no seu PATH no ~/.bashrc ou ~/.zshrc:"
echo '  export PATH="$HOME/.local/bin:$PATH"'
