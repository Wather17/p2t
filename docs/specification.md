# Specification: Framework p2t (Pay to Work)
**Organização Financeira por Exceção e Viabilidade de Renda**

O manifesto em [docs/manifesto.md](manifesto.md) define a filosofia geral do produto. Esta documentação formaliza os módulos matemáticos do projeto `p2t`, utilizando o **método dedutivo direto** para organizar a vida financeira por exceção e mensurar o retorno real do tempo investido no trabalho.

O fechamento financeiro mensal calcula a capacidade de metas e a margem livre a partir da renda recebida, dos compromissos recorrentes e das exceções do período. As equações abaixo permanecem como o módulo de viabilidade da renda trabalhista, não como a definição inteira do produto.

---

## 1. Método 1: Telemetria de Retorno de Tempo ($VRH$ e $IDT$)

### 1.1 Definição de Variáveis
Seja o conjunto de variáveis operacionais de um indivíduo em um período mensal $t$:

* $S_B \in \mathbb{R}^+$: Salário Bruto nominal registrado.
* $D_F \in \mathbb{R}^+$: Somatório dos descontos legais fixos (INSS, IRRF, coparticipações oficiais).
* $D_E \in \mathbb{R}^+$: Descontos decorrentes de erros operacionais (oficiais ou efetuados "por fora").
* $C_I \in \mathbb{R}^+$: Custos invisíveis do bolso para viabilizar o trabalho (transporte não coberto, alimentação extra).
* $H_C \in \mathbb{R}^+$: Horas mensais contratualmente devidas.
* $H_D \in \mathbb{R}^+$: Horas mensais gastas exclusivamente em deslocamento (ida/volta).
* $H_T \in \mathbb{R}^+$: Horas Totais Efetivamente Dedicadas ao Trabalho.

---

### 1.2 Premissas e Axiomas do Modelo
1. **Axioma do Tempo Útil ($H_T$):** O tempo total de vida entregue à atividade laboral engloba obrigatoriamente o deslocamento físico.
   $$H_T = H_C + H_D$$

2. **Axioma da Liquidez Real ($S_L$):** O rendimento líquido disponível real é o salário bruto deduzido de todos os encargos, penalidades e custos operacionais de subsistência no emprego.
   $$S_L = S_B - (D_F + D_E + C_I)$$

---

### 1.3 Modelo de Deslocamento ($H_D$)

$H_D$ pode ser obtido por dois modos:

**A) Estimativa mensal por escala** (sem data de referência):
$$H_D = \text{dias médios da escala} \times \text{horas diárias de deslocamento}$$

| Escala | Dias médios/mês |
|---|---|
| 5x2 | 22 |
| 6x1 | 26 |
| 12x36 | 15 |
| 4x3 | 17 |

**B) Cálculo exato por calendário** (com `--shift-ref-date`): conta os plantões reais do mês de competência e aplica:

$$H_D = \text{plantões} \times \text{horas diárias de deslocamento}$$

Regras de contagem por escala:

* **5x2**: dias de segunda a sexta-feira.
* **6x1**: dias de segunda a sábado.
* **4x3**: dias de segunda a quinta-feira.
* **12x36**: dias $t$ tais que a diferença em dias em relação à data de referência é par: $t_i \equiv 0 \pmod 2$ — ciclo 12h on + 36h off = 48h, plantões a cada 2 dias.

**Convenções do `--shift-ref-date` (escala 12x36):**

* A data informada é o **início do turno** do plantão (não o término).
* O plantão pertence ao mês de competência da **data de início** (turnos que cruzam a meia-noite contam no mês em que começam).
* O conceito é **exclusivo da escala 12x36** — 5x2/6x1/4x3 contam apenas por regra de dias da semana.
* Exemplo: plantão iniciado no dia 5 → os próximos são 7, 9, ... da mesma competência.
* Se a refDate for um dia de folga (ou o término do turno), a paridade inverte e a contagem erra em ±1: informe sempre a data de início.

---

### 1.4 Dedução Lógica das Equações

#### A) Valor Real da Hora Dedicada ($VRH$)
O retorno financeiro real gerado por cada hora de vida dedicada ao trabalho é obtido pela razão entre a Liquidez Real ($S_L$) e a Carga Horária Total ($H_T$):

$$VRH = \frac{S_B - (D_F + D_E + C_I)}{H_C + H_D}$$

#### B) Índice de Desperdício de Trabalho ($IDT$)
A porcentagem do salário bruto corroída exclusivamente por punições operacionais ($D_E$) e custos de permanência no trabalho ($C_I$) é dada por:

$$IDT = \left( \frac{D_E + C_I}{S_B} \right) \times 100$$

---

### 1.5 Matriz de Decisão do $IDT$
A tese de permanência laboral baseia-se no valor de $IDT$ do mês atual ou da sua **Média Móvel de 3 Meses** ($\overline{IDT}_3$):

$$\overline{IDT}_3 = \frac{\sum_{i=0}^{2} IDT_{t-i}}{3}$$

* **$\overline{IDT}_3 < 10\%$:** Zona Verde (Estável / Padrão Operacional).
* **$10\% \le \overline{IDT}_3 < 15\%$:** Zona Amarela (Alerta de Corrosão / Ativar busca passiva de vagas).
* **$\overline{IDT}_3 \ge 15\%$:** Zona Vermelha (Corrosão elevada / considerar um plano de saída).

---

## 2. Método 2: Gestão por Exceção do Buffer Operacional (A Caixinha)

Este método automatiza o cálculo da variável $C_I$ (Custos Invisíveis) utilizando o princípio de telemetria por saldo remanescente.

### 2.1 Definição de Variáveis
* $T \in \mathbb{R}^+$: Teto fixo de liquidez alocado para trabalho (Ex: $R\$ \; 300,00$).
* $S_{rem} \in \mathbb{R}^+$: Saldo não utilizado retido na caixinha ao final do ciclo mensal.
* $R_t \in \mathbb{R}^+$: Valor necessário de reposição no mês $t$.

---

### 2.2 Dedução das Métricas

#### A) Equação de Consumo Real do Ciclo
O custo invisível real ($C_I$) referente ao período $t$ é deduzido pela diferença direta entre o teto estipulado e o saldo remanescente:

$$C_I = R_t = T - S_{rem}$$

#### B) Média Móvel de Reposição ($\bar{R}$)
Para um histórico de $N$ meses registrados:

$$\bar{R} = \frac{1}{N} \sum_{i=1}^{N} R_i$$

#### C) Taxa de Consumo Média ($TCM$)
A proporção de utilização do buffer operacional em relação ao teto total é expressa por:

$$TCM = \left( \frac{\bar{R}}{T} \right) \times 100$$

---

### 2.3 Diagnóstico de Eficiência
* **$TCM < 45\%$:** Alta Eficiência (Otimização de custos de transporte/alimentação).
* **$45\% \le TCM \le 55\%$:** Eficiência Estável (Alinhada às estimativas de projeto).
* **$55\% < TCM \le 60\%$:** Alerta de Consumo Elevado (atenção ao limite crítico).
* **$TCM > 60\%$:** Anomalia de Consumo (Requer auditoria pontual do extrato bancário).
