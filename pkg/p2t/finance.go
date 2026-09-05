package p2t

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// RecurringPeriod representa a periodicidade de um compromisso recorrente.
type RecurringPeriod string

const (
	RecurringMonthly RecurringPeriod = "monthly"
	RecurringYearly  RecurringPeriod = "yearly"
)

// RecurringPurpose representa a finalidade principal do compromisso.
type RecurringPurpose string

const (
	PurposeLife RecurringPurpose = "life"
	PurposeWork RecurringPurpose = "work"
	PurposeGoal RecurringPurpose = "goal"
)

// RecurringEssentiality representa o grau de corte possível do compromisso.
type RecurringEssentiality string

const (
	EssentialityEssential     RecurringEssentiality = "essential"
	EssentialityUseful        RecurringEssentiality = "useful"
	EssentialityDiscretionary RecurringEssentiality = "discretionary"
)

// RecurringStatus representa o ciclo de vida de um compromisso.
type RecurringStatus string

const (
	RecurringActive   RecurringStatus = "active"
	RecurringArchived RecurringStatus = "archived"
)

// Goal representa uma meta financeira acompanhada pelo p2t.
type Goal struct {
	ID                  int64
	Name                string
	TargetAmount        float64
	Deadline            string
	CurrentBalance      float64
	MonthlyContribution float64
	Archived            bool
}

// Validate verifica os dados necessários para uma meta.
func (g Goal) Validate() error {
	if g.Name == "" {
		return errors.New("nome da meta nao pode ser vazio")
	}
	if g.TargetAmount <= 0 {
		return errors.New("valor alvo da meta deve ser maior que zero")
	}
	if g.CurrentBalance < 0 || g.MonthlyContribution < 0 {
		return errors.New("saldo e aporte mensal da meta nao podem ser negativos")
	}
	if err := ValidateReferenceMonth(g.Deadline); err != nil {
		return fmt.Errorf("prazo da meta invalido: %w", err)
	}
	if g.CurrentBalance > g.TargetAmount {
		return errors.New("saldo atual da meta nao pode exceder o valor alvo")
	}
	return nil
}

// ProgressPercent retorna o percentual do valor alvo já acumulado.
func (g Goal) ProgressPercent() float64 {
	if g.TargetAmount <= 0 {
		return 0
	}
	return RoundPercentage((g.CurrentBalance / g.TargetAmount) * 100)
}

// RemainingAmount retorna o valor que falta para atingir a meta.
func (g Goal) RemainingAmount() float64 {
	remaining := g.TargetAmount - g.CurrentBalance
	if remaining < 0 {
		return 0
	}
	return RoundCurrency(remaining)
}

// RecurringCommitment representa uma cobrança recorrente sem exigir o registro de cada transação.
type RecurringCommitment struct {
	ID           int64
	Name         string
	Amount       float64
	Period       RecurringPeriod
	Purpose      RecurringPurpose
	Essentiality RecurringEssentiality
	BillingDay   int
	Status       RecurringStatus
	LastReviewed string
}

// Validate verifica os dados necessários para um compromisso recorrente.
func (r RecurringCommitment) Validate() error {
	if r.Name == "" {
		return errors.New("nome do compromisso nao pode ser vazio")
	}
	if r.Amount <= 0 {
		return errors.New("valor do compromisso deve ser maior que zero")
	}
	if r.Period != RecurringMonthly && r.Period != RecurringYearly {
		return fmt.Errorf("periodicidade invalida: %s", r.Period)
	}
	if r.Purpose != PurposeLife && r.Purpose != PurposeWork && r.Purpose != PurposeGoal {
		return fmt.Errorf("finalidade invalida: %s", r.Purpose)
	}
	if r.Essentiality != EssentialityEssential && r.Essentiality != EssentialityUseful && r.Essentiality != EssentialityDiscretionary {
		return fmt.Errorf("essencialidade invalida: %s", r.Essentiality)
	}
	if r.BillingDay < 0 || r.BillingDay > 31 {
		return errors.New("dia de cobranca deve estar entre 0 e 31")
	}
	if r.Status != RecurringActive && r.Status != RecurringArchived {
		return fmt.Errorf("status invalido: %s", r.Status)
	}
	if r.LastReviewed != "" {
		if err := ValidateReferenceMonth(r.LastReviewed); err != nil {
			return fmt.Errorf("ultima revisao invalida: %w", err)
		}
	}
	return nil
}

// MonthlyEquivalent converte compromissos anuais para um valor mensal comparável.
func (r RecurringCommitment) MonthlyEquivalent() float64 {
	if r.Period == RecurringYearly {
		return RoundCurrency(r.Amount / 12)
	}
	return RoundCurrency(r.Amount)
}

// IsActive informa se o compromisso deve entrar no fechamento mensal.
func (r RecurringCommitment) IsActive() bool {
	return r.Status == RecurringActive
}

// FinancialSnapshot guarda o resumo financeiro de uma competência.
type FinancialSnapshot struct {
	ID                       int64
	ReferenceMonth           string
	ReceivedIncome           float64
	ReserveBalance           float64
	PlannedGoalContributions float64
	ExceptionalCosts         float64
	RecurringCommitmentTotal float64
	GoalCapacity             float64
	FreeMargin               float64
	CreatedAt                time.Time
}

// Validate verifica os dados de um fechamento mensal.
func (s FinancialSnapshot) Validate() error {
	if err := ValidateReferenceMonth(s.ReferenceMonth); err != nil {
		return err
	}
	if s.ReceivedIncome < 0 || s.ReserveBalance < 0 || s.PlannedGoalContributions < 0 || s.ExceptionalCosts < 0 {
		return errors.New("valores financeiros nao podem ser negativos")
	}
	if s.RecurringCommitmentTotal < 0 {
		return errors.New("total de compromissos recorrentes nao pode ser negativo")
	}
	return nil
}

// FinancialSnapshotMetrics contém os resultados deduzidos de um fechamento.
type FinancialSnapshotMetrics struct {
	RecurringCommitmentTotal float64
	GoalCapacity             float64
	FreeMargin               float64
}

// CalculateFinancialSnapshot calcula capacidade de metas e margem livre.
func CalculateFinancialSnapshot(receivedIncome, recurringTotal, plannedContributions, exceptionalCosts float64) (FinancialSnapshotMetrics, error) {
	if receivedIncome < 0 || recurringTotal < 0 || plannedContributions < 0 || exceptionalCosts < 0 {
		return FinancialSnapshotMetrics{}, errors.New("valores financeiros nao podem ser negativos")
	}

	goalCapacity := RoundCurrency(receivedIncome - recurringTotal)
	freeMargin := RoundCurrency(goalCapacity - plannedContributions - exceptionalCosts)
	return FinancialSnapshotMetrics{
		RecurringCommitmentTotal: RoundCurrency(recurringTotal),
		GoalCapacity:             goalCapacity,
		FreeMargin:               freeMargin,
	}, nil
}

var referenceMonthPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// ValidateReferenceMonth valida competências no formato YYYY-MM.
func ValidateReferenceMonth(value string) error {
	if !referenceMonthPattern.MatchString(value) {
		return fmt.Errorf("competencia invalida '%s': esperado YYYY-MM", value)
	}
	return nil
}
