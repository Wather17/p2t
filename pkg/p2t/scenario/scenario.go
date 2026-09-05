// Package scenario executa histórias financeiras determinísticas para validar
// os princípios do p2t como comportamento observável.
package scenario

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/Wather17/p2t/pkg/p2t"
)

const defaultTolerance = 0.01

var knownSignals = map[string]bool{
	"margin_negative":           true,
	"goal_on_track":             true,
	"goal_behind":               true,
	"goal_completed":            true,
	"review_recurring":          true,
	"work_cost_high":            true,
	"work_cash_result_negative": true,
	"work_return_negative":      true,
	"work_corrosion_high":       true,
	"exit_plan_consideration":   true,
}

// Scenario representa uma trajetória financeira descrita em JSON.
type Scenario struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Context   string          `json:"context"`
	Question  string          `json:"question"`
	Tolerance float64         `json:"tolerance"`
	Months    []ScenarioMonth `json:"months"`
}

// ScenarioMonth representa o estado completo de uma competência.
type ScenarioMonth struct {
	ReferenceMonth           string           `json:"reference_month"`
	ReceivedIncome           float64          `json:"received_income"`
	ReserveBalance           float64          `json:"reserve_balance"`
	PlannedGoalContributions float64          `json:"planned_goal_contributions"`
	ExceptionalCosts         float64          `json:"exceptional_costs"`
	Recurring                []RecurringInput `json:"recurring"`
	Goals                    []GoalInput      `json:"goals"`
	Work                     *WorkInput       `json:"work,omitempty"`
	Expect                   Expectations     `json:"expect"`
}

// GoalInput é a representação de uma meta dentro de um cenário.
type GoalInput struct {
	Name                string  `json:"name"`
	TargetAmount        float64 `json:"target_amount"`
	Deadline            string  `json:"deadline"`
	CurrentBalance      float64 `json:"current_balance"`
	MonthlyContribution float64 `json:"monthly_contribution"`
	Archived            bool    `json:"archived"`
}

// RecurringInput é a representação de um compromisso dentro de um cenário.
type RecurringInput struct {
	Name         string                    `json:"name"`
	Amount       float64                   `json:"amount"`
	Period       p2t.RecurringPeriod       `json:"period"`
	Purpose      p2t.RecurringPurpose      `json:"purpose"`
	Essentiality p2t.RecurringEssentiality `json:"essentiality"`
	BillingDay   int                       `json:"billing_day"`
	Status       p2t.RecurringStatus       `json:"status"`
	LastReviewed string                    `json:"last_reviewed"`
}

// WorkInput é a entrada trabalhista usada pelo cálculo de VRH e IDT.
type WorkInput struct {
	GrossSalary     float64 `json:"gross_salary"`
	FixedDeductions float64 `json:"fixed_deductions"`
	ErrorDeductions float64 `json:"error_deductions"`
	InvisibleCosts  float64 `json:"invisible_costs"`
	ContractHours   float64 `json:"contract_hours"`
	CommuteHours    float64 `json:"commute_hours"`
}

// Expectations contém os resultados que a história afirma que o sistema deve produzir.
type Expectations struct {
	RecurringTotal float64            `json:"recurring_total"`
	GoalCapacity   float64            `json:"goal_capacity"`
	FreeMargin     float64            `json:"free_margin"`
	GoalProgress   map[string]float64 `json:"goal_progress"`
	VRH            *float64           `json:"vrh,omitempty"`
	IDT            *float64           `json:"idt,omitempty"`
	Signals        []string           `json:"signals"`
}

// MonthResult contém os resultados calculados para uma competência.
type MonthResult struct {
	ReferenceMonth string
	RecurringTotal float64
	GoalCapacity   float64
	FreeMargin     float64
	GoalProgress   map[string]float64
	VRH            *float64
	IDT            *float64
	Signals        []string
}

// Result contém a trajetória completa calculada.
type Result struct {
	ScenarioID string
	Months     []MonthResult
}

// Failure descreve uma divergência entre uma expectativa e o resultado calculado.
type Failure struct {
	ScenarioID     string
	ReferenceMonth string
	Field          string
	Expected       string
	Actual         string
}

func (f Failure) Error() string {
	return fmt.Sprintf("cenário %s/%s: %s esperado %s, obtido %s", f.ScenarioID, f.ReferenceMonth, f.Field, f.Expected, f.Actual)
}

// Load decodifica e valida um cenário JSON.
func Load(r io.Reader) (Scenario, error) {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	var scenario Scenario
	if err := decoder.Decode(&scenario); err != nil {
		return Scenario{}, fmt.Errorf("falha ao ler cenário: %w", err)
	}
	if err := validateScenario(scenario); err != nil {
		return Scenario{}, err
	}
	if scenario.Tolerance == 0 {
		scenario.Tolerance = defaultTolerance
	}
	return scenario, nil
}

func validateScenario(s Scenario) error {
	if s.ID == "" {
		return errors.New("cenário sem id")
	}
	if s.Title == "" || s.Context == "" || s.Question == "" {
		return fmt.Errorf("cenário %s deve ter title, context e question", s.ID)
	}
	if s.Tolerance < 0 {
		return fmt.Errorf("cenário %s possui tolerância negativa", s.ID)
	}
	if len(s.Months) == 0 {
		return fmt.Errorf("cenário %s não possui competências", s.ID)
	}
	var previousMonth string
	for i, month := range s.Months {
		if err := p2t.ValidateReferenceMonth(month.ReferenceMonth); err != nil {
			return fmt.Errorf("cenário %s mês %d: %w", s.ID, i, err)
		}
		if previousMonth != "" && month.ReferenceMonth <= previousMonth {
			return fmt.Errorf("cenário %s possui competências fora de ordem: %s depois de %s", s.ID, month.ReferenceMonth, previousMonth)
		}
		previousMonth = month.ReferenceMonth
		if month.ReceivedIncome < 0 || month.ReserveBalance < 0 || month.PlannedGoalContributions < 0 || month.ExceptionalCosts < 0 {
			return fmt.Errorf("cenário %s/%s possui valor financeiro negativo", s.ID, month.ReferenceMonth)
		}
		for _, signal := range month.Expect.Signals {
			if !knownSignals[signal] {
				return fmt.Errorf("cenário %s/%s possui sinal desconhecido: %s", s.ID, month.ReferenceMonth, signal)
			}
		}
	}
	return nil
}

// Run calcula todos os meses de uma trajetória sem acessar banco ou terminal.
func Run(s Scenario) (Result, error) {
	if err := validateScenario(s); err != nil {
		return Result{}, err
	}
	if s.Tolerance == 0 {
		s.Tolerance = defaultTolerance
	}

	result := Result{ScenarioID: s.ID, Months: make([]MonthResult, 0, len(s.Months))}
	for _, month := range s.Months {
		monthResult, err := runMonth(s.ID, month, result.Months)
		if err != nil {
			return Result{}, err
		}
		result.Months = append(result.Months, monthResult)
	}
	return result, nil
}

func runMonth(scenarioID string, month ScenarioMonth, history []MonthResult) (MonthResult, error) {
	var recurringTotal float64
	for _, input := range month.Recurring {
		status := input.Status
		if status == "" {
			status = p2t.RecurringActive
		}
		commitment := p2t.RecurringCommitment{
			Name: input.Name, Amount: input.Amount, Period: input.Period,
			Purpose: input.Purpose, Essentiality: input.Essentiality,
			BillingDay: input.BillingDay, Status: status, LastReviewed: input.LastReviewed,
		}
		if err := commitment.Validate(); err != nil {
			return MonthResult{}, fmt.Errorf("cenário %s/%s: compromisso %q inválido: %w", scenarioID, month.ReferenceMonth, input.Name, err)
		}
		if commitment.IsActive() {
			recurringTotal += commitment.MonthlyEquivalent()
		}
	}
	recurringTotal = p2t.RoundCurrency(recurringTotal)
	metrics, err := p2t.CalculateFinancialSnapshot(month.ReceivedIncome, recurringTotal, month.PlannedGoalContributions, month.ExceptionalCosts)
	if err != nil {
		return MonthResult{}, fmt.Errorf("cenário %s/%s: fechamento inválido: %w", scenarioID, month.ReferenceMonth, err)
	}

	goalProgress := make(map[string]float64, len(month.Goals))
	signals := make(map[string]bool)
	for _, input := range month.Goals {
		goal := p2t.Goal{
			Name: input.Name, TargetAmount: input.TargetAmount, Deadline: input.Deadline,
			CurrentBalance: input.CurrentBalance, MonthlyContribution: input.MonthlyContribution,
			Archived: input.Archived,
		}
		if err := goal.Validate(); err != nil {
			return MonthResult{}, fmt.Errorf("cenário %s/%s: meta %q inválida: %w", scenarioID, month.ReferenceMonth, input.Name, err)
		}
		if goal.Archived {
			continue
		}
		goalProgress[goal.Name] = goal.ProgressPercent()
		if goal.CurrentBalance >= goal.TargetAmount {
			signals["goal_completed"] = true
		} else if goalOnTrack(goal, month.ReferenceMonth) {
			signals["goal_on_track"] = true
		} else {
			signals["goal_behind"] = true
		}
	}
	if metrics.FreeMargin < 0 {
		signals["margin_negative"] = true
	}
	for _, input := range month.Recurring {
		status := input.Status
		if status == "" {
			status = p2t.RecurringActive
		}
		commitment := p2t.RecurringCommitment{
			Name: input.Name, Amount: input.Amount, Period: input.Period,
			Purpose: input.Purpose, Essentiality: input.Essentiality,
			BillingDay: input.BillingDay, Status: status, LastReviewed: input.LastReviewed,
		}
		if commitment.IsActive() && commitment.Essentiality == p2t.EssentialityDiscretionary && staleReview(commitment.LastReviewed, month.ReferenceMonth) {
			signals["review_recurring"] = true
		}
	}

	var vrh, idt *float64
	if month.Work != nil {
		telemetry, err := p2t.CalculateTelemetry(p2t.TelemetryInput{
			GrossSalary: month.Work.GrossSalary, FixedDeductions: month.Work.FixedDeductions,
			ErrorDeductions: month.Work.ErrorDeductions, InvisibleCosts: month.Work.InvisibleCosts,
			ContractHours: month.Work.ContractHours, CommuteHours: month.Work.CommuteHours,
		})
		if err != nil {
			return MonthResult{}, fmt.Errorf("cenário %s/%s: telemetria inválida: %w", scenarioID, month.ReferenceMonth, err)
		}
		vrhValue, idtValue := telemetry.VRH, telemetry.IDT
		vrh, idt = &vrhValue, &idtValue
		if idtValue >= p2t.ZoneIDTGreenThreshold {
			signals["work_cost_high"] = true
		}
		if idtValue >= p2t.ZoneIDTRedThreshold {
			signals["work_corrosion_high"] = true
		}
		if telemetry.RealLiquidity < 0 {
			signals["work_cash_result_negative"] = true
		}
		if telemetry.VRH <= 0 {
			signals["work_return_negative"] = true
		}
		if persistentWorkDeterioration(history, telemetry) {
			signals["exit_plan_consideration"] = true
		}
	}

	return MonthResult{
		ReferenceMonth: month.ReferenceMonth,
		RecurringTotal: metrics.RecurringCommitmentTotal,
		GoalCapacity:   metrics.GoalCapacity,
		FreeMargin:     metrics.FreeMargin,
		GoalProgress:   goalProgress,
		VRH:            vrh, IDT: idt,
		Signals: sortedSignals(signals),
	}, nil
}

func goalOnTrack(goal p2t.Goal, referenceMonth string) bool {
	monthsRemaining := monthsInclusive(referenceMonth, goal.Deadline)
	return monthsRemaining > 0 && goal.MonthlyContribution*float64(monthsRemaining) >= goal.RemainingAmount()
}

func monthsInclusive(from, to string) int {
	fromDate, errFrom := time.Parse("2006-01", from)
	toDate, errTo := time.Parse("2006-01", to)
	if errFrom != nil || errTo != nil || toDate.Before(fromDate) {
		return 0
	}
	return (toDate.Year()-fromDate.Year())*12 + int(toDate.Month()-fromDate.Month()) + 1
}

func staleReview(lastReviewed, referenceMonth string) bool {
	if lastReviewed == "" {
		return true
	}
	last, errLast := time.Parse("2006-01", lastReviewed)
	current, errCurrent := time.Parse("2006-01", referenceMonth)
	if errLast != nil || errCurrent != nil || current.Before(last) {
		return false
	}
	months := (current.Year()-last.Year())*12 + int(current.Month()-last.Month())
	return months > 6
}

func persistentWorkDeterioration(history []MonthResult, current p2t.TelemetryResult) bool {
	if len(history) < 2 || current.IDT < p2t.ZoneIDTRedThreshold {
		return false
	}
	first := history[len(history)-2]
	second := history[len(history)-1]
	if first.IDT == nil || second.IDT == nil || first.VRH == nil || *first.IDT < p2t.ZoneIDTRedThreshold || *second.IDT < p2t.ZoneIDTRedThreshold {
		return false
	}
	return current.VRH < *first.VRH
}

func sortedSignals(signals map[string]bool) []string {
	result := make([]string, 0, len(signals))
	for signal, active := range signals {
		if active {
			result = append(result, signal)
		}
	}
	sort.Strings(result)
	return result
}

// Compare compara resultados calculados com as expectativas do cenário.
func Compare(s Scenario, result Result) []Failure {
	tolerance := s.Tolerance
	if tolerance == 0 {
		tolerance = defaultTolerance
	}
	var failures []Failure
	for i, month := range s.Months {
		if i >= len(result.Months) {
			failures = append(failures, Failure{ScenarioID: s.ID, ReferenceMonth: month.ReferenceMonth, Field: "month", Expected: "resultado presente", Actual: "ausente"})
			continue
		}
		actual := result.Months[i]
		expected := month.Expect
		failures = appendNumericFailure(failures, s.ID, month.ReferenceMonth, "recurring_total", expected.RecurringTotal, actual.RecurringTotal, tolerance)
		failures = appendNumericFailure(failures, s.ID, month.ReferenceMonth, "goal_capacity", expected.GoalCapacity, actual.GoalCapacity, tolerance)
		failures = appendNumericFailure(failures, s.ID, month.ReferenceMonth, "free_margin", expected.FreeMargin, actual.FreeMargin, tolerance)
		for name, value := range expected.GoalProgress {
			actualValue, ok := actual.GoalProgress[name]
			if !ok {
				failures = append(failures, Failure{ScenarioID: s.ID, ReferenceMonth: month.ReferenceMonth, Field: "goal_progress." + name, Expected: fmt.Sprintf("%.2f", value), Actual: "ausente"})
				continue
			}
			failures = appendNumericFailure(failures, s.ID, month.ReferenceMonth, "goal_progress."+name, value, actualValue, tolerance)
		}
		if expected.VRH != nil && (actual.VRH == nil || !within(*expected.VRH, *actual.VRH, tolerance)) {
			actualValue := "ausente"
			if actual.VRH != nil {
				actualValue = fmt.Sprintf("%.2f", *actual.VRH)
			}
			failures = append(failures, Failure{ScenarioID: s.ID, ReferenceMonth: month.ReferenceMonth, Field: "vrh", Expected: fmt.Sprintf("%.2f", *expected.VRH), Actual: actualValue})
		}
		if expected.IDT != nil && (actual.IDT == nil || !within(*expected.IDT, *actual.IDT, tolerance)) {
			actualValue := "ausente"
			if actual.IDT != nil {
				actualValue = fmt.Sprintf("%.2f", *actual.IDT)
			}
			failures = append(failures, Failure{ScenarioID: s.ID, ReferenceMonth: month.ReferenceMonth, Field: "idt", Expected: fmt.Sprintf("%.2f", *expected.IDT), Actual: actualValue})
		}
		if expected.Signals != nil && !sameStrings(expected.Signals, actual.Signals) {
			failures = append(failures, Failure{ScenarioID: s.ID, ReferenceMonth: month.ReferenceMonth, Field: "signals", Expected: fmt.Sprintf("%v", expected.Signals), Actual: fmt.Sprintf("%v", actual.Signals)})
		}
	}
	return failures
}

func appendNumericFailure(failures []Failure, scenarioID, month, field string, expected, actual, tolerance float64) []Failure {
	if !within(expected, actual, tolerance) {
		return append(failures, Failure{ScenarioID: scenarioID, ReferenceMonth: month, Field: field, Expected: fmt.Sprintf("%.2f", expected), Actual: fmt.Sprintf("%.2f", actual)})
	}
	return failures
}

func within(expected, actual, tolerance float64) bool {
	delta := expected - actual
	if delta < 0 {
		delta = -delta
	}
	return delta <= tolerance
}

func sameStrings(expected, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}
	expectedCopy := append([]string(nil), expected...)
	actualCopy := append([]string(nil), actual...)
	sort.Strings(expectedCopy)
	sort.Strings(actualCopy)
	for i := range expectedCopy {
		if expectedCopy[i] != actualCopy[i] {
			return false
		}
	}
	return true
}
