# Manifesto do p2t

## O problema que o p2t resolve

O p2t é um sistema de organização financeira pessoal por exceção. Ele não tenta substituir o extrato bancário nem exige o registro de cada centavo gasto. O extrato guarda o detalhe do passado; o p2t transforma os fatos importantes em decisões sobre o presente e o futuro.

O sistema existe para responder perguntas práticas:

1. Minha renda sustenta meus compromissos e minhas metas?
2. Quanto consigo direcionar para o que importa?
3. Quais cobranças continuam acontecendo sem revisão?
4. Quanto custa produzir minha renda?
5. O trabalho ainda entrega retorno suficiente pelo dinheiro e pelo tempo consumidos?

## Princípios

### 1. Registro por decisão, não por transação

O p2t registra fechamentos mensais, metas, reservas, compromissos recorrentes e exceções. A categorização detalhada de compras continua sendo responsabilidade do extrato ou de outra ferramenta.

### 2. A renda precisa deixar capacidade

A organização começa pela renda efetivamente recebida e pelos compromissos recorrentes conhecidos:

$$\text{Capacidade de Metas} = \text{Renda Recebida} - \text{Compromissos Recorrentes}$$

Depois dos aportes planejados e dos custos excepcionais do mês, o que resta é a margem livre:

$$\text{Margem Livre} = \text{Capacidade de Metas} - \text{Aportes Planejados} - \text{Custos Excepcionais}$$

Uma margem negativa não é uma falha moral; é um sinal de que o plano atual não cabe na renda observada.

### 3. Metas tornam a sobra intencional

Uma meta tem nome, valor-alvo, prazo, saldo atual e aporte mensal planejado. O sistema mede progresso e capacidade de financiamento, sem impor uma taxonomia rígida de objetivos.

### 4. Compromissos recorrentes são vazamentos silenciosos

Assinaturas, aluguel, parcelas, seguros e mensalidades devem ser acompanhados como compromissos recorrentes. Cada item possui periodicidade, finalidade, essencialidade, data de cobrança e data de revisão.

Valores anuais são convertidos para equivalentes mensais para que o impacto real fique visível.

### 5. O trabalho é uma fonte de renda que precisa se justificar

A análise trabalhista permanece importante, mas é um módulo do cockpit financeiro. Ela acompanha o custo direto de produzir renda, o tempo dedicado e o retorno por hora.

O custo direto inicial do trabalho é:

$$C_{trabalho} = D_E + C_I$$

O retorno por hora continua sendo:

$$VRH = \frac{S_B - (D_F + D_E + C_I)}{H_C + H_D}$$

O VRH e o IDT não são o produto inteiro. São instrumentos para perceber corrosão persistente e decidir quando iniciar um plano de saída ou mudança de renda.

### 6. Sinais orientam; não decidem sozinhos

Um mês ruim não determina uma demissão. Os sinais devem ser interpretados junto com margem financeira, reservas, metas, tendência histórica e alternativas disponíveis. O p2t exibe alertas explicáveis, não ordens automáticas.

### 7. Cenários são TDD filosófico

Os princípios do p2t devem ser testados como histórias financeiras, não apenas como fórmulas isoladas. Um cenário descreve uma pessoa fictícia ao longo de vários meses, declara o que deveria ser percebido e transforma essa expectativa em um teste executável.

O ciclo é:

1. Escrever a história e a pergunta decisória.
2. Definir métricas e sinais esperados.
3. Rodar a história contra o modelo atual.
4. Implementar a menor mudança necessária.
5. Preservar os cenários anteriores como contratos do produto.

Os cenários canônicos ficam em `pkg/p2t/scenario/testdata/` e validam números e decisões. Eles são fixos, determinísticos e não substituem o extrato nem tentam representar uma verdade universal sobre como cada pessoa deve viver.

## Fechamento mensal

O fechamento mensal é a unidade principal de acompanhamento. Ele registra apenas:

- renda recebida;
- saldo total das reservas;
- compromissos recorrentes ativos;
- aportes planejados para metas;
- custos excepcionais;
- progresso das metas;
- referência opcional à telemetria do trabalho da mesma competência.

O fechamento pode ser refeito para a mesma competência. O último resumo substitui o anterior; o histórico de progresso é preservado entre competências diferentes.

## O que está fora do escopo

- registrar cada compra ou cobrança individual;
- substituir o extrato bancário;
- criar um plano contábil detalhado;
- decidir automaticamente pela saída de um emprego;
- tratar impostos como custo incremental do trabalho sem explicação;
- transformar metas pessoais em categorias obrigatórias.

## Vocabulário do produto

- **Cockpit financeiro:** visão geral de renda, compromissos, metas, reservas e sinais.
- **Fechamento:** resumo mensal de uma competência `YYYY-MM`.
- **Meta:** objetivo financeiro com alvo, prazo e progresso.
- **Compromisso recorrente:** cobrança que se repete e precisa ser revisada.
- **Capacidade de metas:** valor disponível depois dos compromissos recorrentes.
- **Margem livre:** valor restante depois de aportes e custos excepcionais.
- **Telemetria do trabalho:** módulo que mede o retorno financeiro e temporal da renda laboral.
