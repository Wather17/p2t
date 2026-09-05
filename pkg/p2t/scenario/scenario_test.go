package scenario_test

import (
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wather17/p2t/pkg/p2t/scenario"
)

//go:embed testdata/*.json
var scenarioFiles embed.FS

func TestCanonicalScenarios(t *testing.T) {
	paths, err := fs.Glob(scenarioFiles, "testdata/*.json")
	if err != nil {
		t.Fatalf("falha ao listar cenários: %v", err)
	}
	if len(paths) != 4 {
		t.Fatalf("esperados 4 cenários canônicos, encontrados %d", len(paths))
	}

	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			file, err := scenarioFiles.Open(path)
			if err != nil {
				t.Fatalf("falha ao abrir cenário: %v", err)
			}
			defer file.Close()

			definition, err := scenario.Load(file)
			if err != nil {
				t.Fatalf("falha ao carregar cenário: %v", err)
			}
			result, err := scenario.Run(definition)
			if err != nil {
				t.Fatalf("falha ao executar cenário: %v", err)
			}
			for _, failure := range scenario.Compare(definition, result) {
				t.Error(failure)
			}
		})
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	_, err := scenario.Load(strings.NewReader(`{"id":"broken","title":"x","context":"x","question":"x","unknown":true,"months":[]}`))
	if err == nil {
		t.Fatal("esperado erro para campo JSON desconhecido")
	}
}

func TestLoadRejectsInvalidMonth(t *testing.T) {
	_, err := scenario.Load(strings.NewReader(`{"id":"broken","title":"x","context":"x","question":"x","months":[{"reference_month":"2026-13"}]}`))
	if err == nil {
		t.Fatal("esperado erro para competência inválida")
	}
}

func TestLoadRejectsUnknownSignalAndOutOfOrderMonths(t *testing.T) {
	_, err := scenario.Load(strings.NewReader(`{"id":"broken","title":"x","context":"x","question":"x","months":[{"reference_month":"2026-02","expect":{"signals":["unknown"]}},{"reference_month":"2026-01","expect":{"signals":[]}}]}`))
	if err == nil {
		t.Fatal("esperado erro para sinal desconhecido ou competências fora de ordem")
	}
}
