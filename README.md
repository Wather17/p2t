# p2t (Pay to Work)

> **Organização Financeira por Exceção e Viabilidade de Renda**

O **p2t** é um cockpit financeiro pessoal baseado em fechamentos mensais, metas, reservas e compromissos recorrentes. Ao contrário dos aplicativos tradicionais de organização financeira focados em registrar cada despesa, o p2t usa o extrato como fonte de detalhe e acompanha apenas decisões, exceções e sinais importantes.

A filosofia completa está em [docs/manifesto.md](docs/manifesto.md). A análise do trabalho continua existindo como um módulo para medir o retorno real da renda e perceber quando pode ser necessário planejar uma mudança.

---

## 📐 Fundamentação Matemática

A documentação matemática completa com a dedução das equações está disponível em [docs/specification.md](docs/specification.md).

### Principais Métricas do Trabalho

1. **VRH (Valor Real da Hora Dedicada)**: Retorno financeiro real por hora de vida dedicada ao trabalho (incluindo tempo de deslocamento).
   $$VRH = \frac{S_B - (D_F + D_E + C_I)}{H_C + H_D}$$

2. **IDT (Índice de Desperdício de Trabalho)**: Porcentagem do salário bruto corroída por punições operacionais e custos de permanência.
   $$IDT = \left( \frac{D_E + C_I}{S_B} \right) \times 100$$

3. **Matriz de Decisão ($\overline{IDT}_3$)**: Média móvel de 3 meses do IDT para avaliação de viabilidade laboral:
   * **$\overline{IDT}_3 < 10\%$**: **Zona Verde** (Estável / Padrão Operacional).
   * **$10\% \le \overline{IDT}_3 < 15\%$**: **Zona Amarela** (Alerta de Corrosão / Ativar busca passiva de vagas).
   * **$\overline{IDT}_3 \ge 15\%$**: **Zona Vermelha** (Corrosão elevada / considerar um plano de saída).

4. **Buffer Operacional / A Caixinha ($TCM$)**: Gestão por exceção do buffer de liquidez para custos de permanência e transporte.
   $$TCM = \left( \frac{\bar{R}}{T} \right) \times 100$$

### Fechamento Financeiro

O fechamento mensal não exige o registro de cada transação. Ele calcula a capacidade de financiar metas e a margem livre:

$$\text{Capacidade de Metas} = \text{Renda Recebida} - \text{Compromissos Recorrentes}$$

$$\text{Margem Livre} = \text{Capacidade de Metas} - \text{Aportes Planejados} - \text{Custos Excepcionais}$$

---

## 🚀 Instalação e Compilação

### Pré-requisitos
* Go 1.22+ instalado.

### Compilação Local

```bash
# Clonar o repositório
git clone https://github.com/Wather17/p2t.git
cd p2t

# Compilar o binário
go build -o bin/p2t ./cmd/p2t
```

---

## 💻 Como Usar (CLI)

### 1. Telemetria de Retorno de Tempo (`p2t telemetry`)

Calcula o $VRH$, $IDT$, consulta o histórico no SQLite e gera o diagnóstico da Matriz de Decisão:

```bash
./bin/p2t telemetry \
  -s 5000 \
  -H 160 \
  -d 40 \
  -f 800 \
  -e 200 \
  -c 300
```

#### Flags Disponíveis:
* `-s, --gross-salary`: Salário Bruto nominal ($S_B$) **[Obrigatório]**.
* `-H, --contract-hours`: Carga Horária Contratual em horas ($H_C$) **[Obrigatório]**.
* `-d, --commute-hours`: Horas de Deslocamento mensal ($H_D$) manual.
* `-W, --schedule`: Escala de trabalho presencial (5x2, 6x1, 12x36, 4x3) — requer `-D`.
* `-D, --daily-commute`: Horas diárias de deslocamento (ida + volta) para cálculo do HD via escala (ex.: `1.5`).
* `-R, --shift-ref-date`: Data **de início** de um plantão 12x36 (`YYYY-MM-DD`); exclusiva do 12x36 — ativa o cálculo **exato por calendário**.
* `-m, --reference-month`: Mês de competência (`YYYY-MM`; padrão: mês anterior).
* `-I, --interactive`: Modo interativo (formulário guiado).
* `-f, --fixed-deductions`: Descontos Legais Fixos ($D_F$: INSS, IRRF).
* `-e, --error-deductions`: Descontos por erros operacionais ($D_E$).
* `-c, --invisible-costs`: Custos invisíveis do bolso ($C_I$).
* `-i, --idt-history`: Histórico manual de IDTs separados por vírgula (ex: `8.5,9.0`).
* `--db`: Caminho do banco SQLite personalizado (padrão: `~/.p2t/p2t.db`).
* `--no-save`: Executa sem persistir o registro no SQLite.

> **Prioridade do cálculo de HD:** com `-R` (e `-W 12x36`), o HD é calculado de forma **exata** pelo calendário do mês de competência (plantões do mês × horas diárias). Sem `-R`, `-W` + `-D` usa a **estimativa mensal** da escala (dias médios: 5x2=22, 6x1=26, 12x36=15, 4x3=17). Se `-d` for informado junto com escala, o valor calculado prevalece (aviso na saída).

---

### 2. Gestão do Buffer Operacional (`p2t buffer`)

Calcula os custos invisíveis reais ($C_I$), atualiza o histórico e gera o diagnóstico da Taxa de Consumo Média ($TCM$):

```bash
./bin/p2t buffer \
  -t 300 \
  -r 180
```

#### Flags Disponíveis:
* `-t, --cap`: Teto fixo alocado para o trabalho ($T$) **[Obrigatório]**.
* `-r, --remaining`: Saldo remanescente retido na caixinha ($S_{rem}$).
* `-p, --replenishments`: Histórico manual de reposições passadas (ex: `120,150,110`).
* `--db`: Caminho do banco SQLite personalizado (padrão: `~/.p2t/p2t.db`).
* `--no-save`: Executa sem persistir o registro no SQLite.

#### Faixas de Diagnóstico do TCM:
* **TCM < 45%:** Alta Eficiência.
* **45% ≤ TCM ≤ 55%:** Eficiência Estável.
* **55% < TCM ≤ 60%:** Alerta de Consumo Elevado (atenção ao limite crítico).
* **TCM > 60%:** Anomalia de Consumo (auditoria do extrato).

---

### 3. Versão

```bash
./bin/p2t version
```

### 4. Metas Financeiras (`p2t goal`)

Metas são acompanhadas por valor-alvo, prazo, saldo atual e aporte mensal planejado:

```bash
./bin/p2t goal add \
  --name "Reserva de emergência" \
  --target 10000 \
  --deadline 2027-12 \
  --balance 2500 \
  --monthly-contribution 500

./bin/p2t goal list
```

### 5. Compromissos Recorrentes (`p2t recurring`)

Assinaturas, parcelas, seguros e mensalidades são cadastrados sem lançar cada cobrança:

```bash
./bin/p2t recurring add \
  --name "Serviço de música" \
  --amount 25 \
  --period monthly \
  --purpose life \
  --essentiality discretionary \
  --billing-day 10

./bin/p2t recurring list
```

Periodicidades anuais são convertidas automaticamente para equivalente mensal.

### 6. Fechamento Financeiro (`p2t close`)

O comando guiado consolida renda, reservas, metas, compromissos e exceções. Também resume a telemetria trabalhista da mesma competência quando existir:

```bash
./bin/p2t close --interactive --reference-month 2026-08
```

O mesmo mês pode ser fechado novamente; o resumo anterior é atualizado em vez de duplicado.

---

## 🗄️ Armazenamento (SQLite)

O **p2t** utiliza um banco de dados **SQLite leve e local** (driver CGO-free `modernc.org/sqlite`).
Por padrão, os dados são salvos em `~/.p2t/p2t.db`.

---

## 🧪 Testes

Para rodar a suíte completa de testes unitários e de integração:

```bash
go test -v ./...
```

A suíte também executa cenários financeiros multi-mês que validam a filosofia do produto, como estabilidade, crescimento de assinaturas, pagamento para trabalhar e deterioração persistente do retorno laboral.

---

## 📜 Licença

Projeto desenvolvido sob licença open-source MIT.
