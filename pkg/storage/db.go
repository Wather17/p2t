package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenDB abre ou cria uma conexao com o banco SQLite no caminho fornecido (ou em memoria se path == ":memory:").
func OpenDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("falha ao obter diretorio home do usuario: %w", err)
		}
		dir := filepath.Join(home, ".p2t")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("falha ao criar diretorio de dados: %w", err)
		}
		dbPath = filepath.Join(dir, "p2t.db")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir banco sqlite em %s: %w", dbPath, err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("falha na migracao do banco de dados: %w", err)
	}

	return db, nil
}

type migration struct {
	version int
	stmt    string
}

var migrations = []migration{
	{
		version: 1,
		stmt: `
		CREATE TABLE IF NOT EXISTS telemetry_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			gross_salary REAL NOT NULL,
			fixed_deductions REAL NOT NULL,
			error_deductions REAL NOT NULL,
			invisible_costs REAL NOT NULL,
			contract_hours REAL NOT NULL,
			commute_hours REAL NOT NULL,
			total_hours REAL NOT NULL,
			real_liquidity REAL NOT NULL,
			vrh REAL NOT NULL,
			idt REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS buffer_cycles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			cap REAL NOT NULL,
			remaining_balance REAL NOT NULL,
			invisible_cost REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`,
	},
	{
		version: 2,
		stmt: `
		ALTER TABLE telemetry_cycles ADD COLUMN reference_month TEXT NOT NULL DEFAULT '';
		CREATE UNIQUE INDEX IF NOT EXISTS idx_telemetry_reference_month ON telemetry_cycles(reference_month) WHERE reference_month != '';
		`,
	},
}

// GetUserVersion retorna a versao atual do esquema do banco SQLite.
func GetUserVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("PRAGMA user_version;").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("falha ao ler PRAGMA user_version: %w", err)
	}
	return version, nil
}

func migrate(db *sql.DB) error {
	currentVersion, err := GetUserVersion(db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.version > currentVersion {
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("falha ao iniciar transacao para migracao v%d: %w", m.version, err)
			}

			if _, err := tx.Exec(m.stmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("falha na migracao v%d: %w", m.version, err)
			}

			pragmaStmt := fmt.Sprintf("PRAGMA user_version = %d;", m.version)
			if _, err := tx.Exec(pragmaStmt); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("falha ao atualizar user_version para v%d: %w", m.version, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("falha ao commitar migracao v%d: %w", m.version, err)
			}
		}
	}

	return nil
}

