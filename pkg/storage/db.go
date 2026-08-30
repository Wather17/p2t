package storage

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	// dbDirMode restringe o diretorio de dados ao usuario dono.
	dbDirMode = 0o700
	// dbFileMode restringe o arquivo de banco (dados financeiros) ao usuario dono.
	dbFileMode = 0o600
)

// OpenDB abre ou cria uma conexao com o banco SQLite no caminho fornecido (ou em memoria se path == ":memory:").
// O diretorio de dados padrao (~/.p2t) e criado com permissao 0700 e o arquivo de banco e restringido a 0600,
// corrigindo de forma idempotente bancos legados criados com permissao aberta.
//
// Bancos em arquivo abrem com busy_timeout(5000) (espera ate 5s em vez de falhar com "database is locked"
// no uso concorrente, ex.: TUI mais CLI em terminais distintos) e journal_mode(WAL).
func OpenDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("falha ao obter diretorio home do usuario: %w", err)
		}
		dir := filepath.Join(home, ".p2t")
		if err := os.MkdirAll(dir, dbDirMode); err != nil {
			return nil, fmt.Errorf("falha ao criar diretorio de dados: %w", err)
		}
		dbPath = filepath.Join(dir, "p2t.db")
	}

	dsn := dbPath
	if dbPath != ":memory:" {
		q := url.Values{}
		q.Add("_pragma", "busy_timeout(5000)")
		q.Add("_pragma", "journal_mode(WAL)")
		dsn = dbPath + "?" + q.Encode()
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir banco sqlite em %s: %w", dbPath, err)
	}

	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("falha na migracao do banco de dados: %w", err)
	}

	if dbPath != ":memory:" {
		if err := os.Chmod(dbPath, dbFileMode); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("falha ao restringir permissoes do banco %s para 0600: %w", dbPath, err)
		}
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
	{
		version: 3,
		stmt: `
		ALTER TABLE buffer_cycles ADD COLUMN reference_month TEXT NOT NULL DEFAULT '';
		CREATE UNIQUE INDEX IF NOT EXISTS idx_buffer_reference_month ON buffer_cycles(reference_month) WHERE reference_month != '';
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
