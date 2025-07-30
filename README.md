# Projeto: Barbearia Concorrente

Disciplina: Paradigmas de Programação  
Professor: Sidney Nogueira  
Alunos: [Nome dos integrantes do grupo]

---

## Descrição do Projeto

Este projeto simula o funcionamento de uma barbearia utilizando programação concorrente em Go. O objetivo é aplicar conceitos de goroutines, canais, mutex e atomicidade no controle de múltiplos clientes e barbeiros em um ambiente compartilhado.

---

## Funcionamento Geral

- A barbearia possui 2 barbeiros e 4 cadeiras de espera.
- Clientes chegam periodicamente e tomam decisões com base na situação:
  - Sentam-se se houver lugar disponível.
  - Desistem se a fila estiver cheia.
  - Desistem ao ver todos os barbeiros dormindo.
  - Também podem desistir espontaneamente da fila, com chance aleatória.

- Barbeiros trabalham em paralelo:
  - Dormem quando não há clientes.
  - Acordam ao receber um cliente.
  - Cortam o cabelo dos clientes com duração aleatória.

---

## Entidades Concorrentes

- **Barbeiros**: executados como goroutines, verificam e atendem a fila de espera.
- **Clientes**: gerados ciclicamente na função `main`, interagem com o sistema.
- **Monitor de estatísticas**: goroutine que imprime periodicamente os dados da execução.
- **Goroutine de desistência aleatória**: retira clientes da fila com probabilidade definida.

---

## Recursos Compartilhados

- **Fila de espera**: canal limitado (`chan int`) representando as cadeiras.
- **Lista de barbeiros**: estrutura compartilhada para checar o estado (acordado ou dormindo).
- **Contadores de estatísticas**: variáveis globais atualizadas por várias goroutines.

---

## Mecanismos de Controle de Concorrência

- **Canais (channels)**: controle da fila de espera.
- **Mutex (`sync.Mutex`)**: evita que múltiplos barbeiros atendam simultaneamente sem controle.
- **Operações atômicas (`sync/atomic`)**: garantem atualização segura das estatísticas.
- **Funções aleatórias e pausas (`rand`, `time.Sleep`)**: simulam comportamento realista.

---

## Parametrização

Parâmetros configuráveis no início do código:

- `numBarbeiros`: número de barbeiros.
- `numCadeiras`: número de cadeiras de espera.
- `chanceSono`: probabilidade de um barbeiro dormir.
- `chanceDesistir`: probabilidade de um cliente sair da fila.

Todos os parâmetros podem ser alterados facilmente por constantes globais.

---

## Propriedades Atendidas

- **Liveness**: o sistema sempre tenta atender os clientes enquanto houver barbeiros disponíveis.
- **Starvation**: evitada com a utilização de canais, garantindo ordem de atendimento.
- **Race conditions**: prevenidas com o uso de mutex e operações atômicas.

---

## Testes Realizados

O sistema foi testado com os seguintes cenários:

- Número padrão de 2 barbeiros e 4 cadeiras.
- Testes com chegada de clientes a cada 2 segundos.
- Cortes de cabelo simulados com duração entre 2 e 3 segundos.
- Verificação do comportamento ao aumentar/descer os parâmetros de chance de desistência.
- Observação das estatísticas impressas a cada 5 segundos para validar o comportamento geral.

---

## Uso de Código Existente ou IA

O código foi desenvolvido exclusivamente para este projeto.  
Este arquivo README foi redigido com apoio de ferramenta de IA para organização textual e estruturação.

---

## Instruções de Execução

1. Instale o Go: https://go.dev/dl/
2. Execute o projeto com o seguinte comando:

```bash
go run barbeiro.go
