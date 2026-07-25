# Specification: Framework p2t (Pay to Work)
**Modelagem Matemática e Telemetria de Eficiência de Trabalho**

Esta documentação formaliza a fundamentação matemática do projeto `p2t`, utilizando o **método dedutivo direto** para mensurar o retorno real do tempo investido e detectar perdas financeiras operacionais no ambiente de trabalho.

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

### 1.3 Dedução Lógica das Equações

#### A) Valor Real da Hora Dedicada ($VRH$)
O retorno financeiro real gerado por cada hora de vida dedicada ao trabalho é obtido pela razão entre a Liquidez Real ($S_L$) e a Carga Horária Total ($H_T$):

$$VRH = \frac{S_B - (D_F + D_E + C_I)}{H_C + H_D}$$

#### B) Índice de Desperdício de Trabalho ($IDT$)
A porcentagem do salário bruto corroída exclusivamente por punições operacionais ($D_E$) e custos de permanência no trabalho ($C_I$) é dada por:

$$IDT = \left( \frac{D_E + C_I}{S_B} \right) \times 100$$

---

### 1.4 Matriz de Decisão do $IDT$
A tese de permanência laboral baseia-se no valor de $IDT$ do mês atual ou da sua **Média Móvel de 3 Meses** ($\overline{IDT}_3$):

$$\overline{IDT}_3 = \frac{\sum_{i=0}^{2} IDT_{t-i}}{3}$$

* **$\overline{IDT}_3 < 10\%$:** Zona Verde (Estável / Padrão Operacional).
* **$10\% \le \overline{IDT}_3 < 15\%$:** Zona Amarela (Alerta de Corrosão / Ativar busca passiva de vagas).
* **$\overline{IDT}_3 \ge 15\%$:** Zona Vermelha (Inviabilidade Financeira / Gatilho de Saída Ativado).

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
* **$TCM > 60\%$:** Anomalia de Consumo (Requer auditoria pontual do extrato bancário).
