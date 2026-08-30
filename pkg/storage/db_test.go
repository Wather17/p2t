package storage_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Wather17/p2t/pkg/storage"
)

func TestOpenDB_MigrationsAndVersion(t *testing.T) {
	db, err := storage.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("falha ao abrir banco em memoria: %v", err)
	}
	defer db.Close()

	version, err := storage.GetUserVersion(db)
	if err != nil {
		t.Fatalf("falha ao consultar versao do esquema: %v", err)
	}

	if version < 1 {
		t.Errorf("esperado user_version >= 1 apos migracao, obtido: %d", version)
	}
}

func filePerm(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("falha ao stat %s: %v", path, err)
	}
	return info.Mode().Perm()
}

func TestOpenDB_FilePermissionsUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("teste de permissao Unix indisponivel em windows")
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "p2t.db")

	db, err := storage.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("falha ao abrir banco em %s: %v", dbPath, err)
	}
	defer db.Close()

	if got := filePerm(t, dbPath); got != 0o600 {
		t.Errorf("permissoes do banco = %o, esperado 600", got)
	}
}

func TestOpenDB_DefaultDirPermissionsUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("teste de permissao Unix indisponivel em windows")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)

	db, err := storage.OpenDB("")
	if err != nil {
		t.Fatalf("falha ao abrir banco default: %v", err)
	}
	defer db.Close()

	dir := filepath.Join(home, ".p2t")
	if got := filePerm(t, dir); got != 0o700 {
		t.Errorf("permissoes do diretorio = %o, esperado 700", got)
	}
	if got := filePerm(t, filepath.Join(dir, "p2t.db")); got != 0o600 {
		t.Errorf("permissoes do banco = %o, esperado 600", got)
	}
}

func TestOpenDB_CorrectsLegacyPermissionsUnix(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("teste de permissao Unix indisponivel em windows")
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "p2t.db")

	db, err := storage.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("falha ao abrir banco em %s: %v", dbPath, err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("falha ao fechar banco: %v", err)
	}

	if err := os.Chmod(dbPath, 0o644); err != nil {
		t.Fatalf("falha ao simular permissao legada 0644: %v", err)
	}

	db, err = storage.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("falha ao reabrir banco: %v", err)
	}
	defer db.Close()

	if got := filePerm(t, dbPath); got != 0o600 {
		t.Errorf("permissoes do banco apos correcao = %o, esperado 600", got)
	}
}
