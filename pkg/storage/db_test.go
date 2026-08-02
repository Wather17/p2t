package storage_test

import (
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
