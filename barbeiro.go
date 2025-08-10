package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const ( // contantes
	numCadeiras    = 4
	numBarbeiros   = 2
	chanceSono     = 30
	chanceDesistir = 30
)

var ( //variaveis
	cadeiras           = make(chan int, numCadeiras) //make permite a criação de tipos especificos de dados como o chan/ chanel / canal
	mu                 sync.Mutex                    //criação do mutex
	barbeiros          []*Barbeiro                   //lista para verificação se ambos estão dormindo
	atendidos          int32                         // total atendidos
	desistiramFila     int32                         // cliente desistiu depois de entrar
	desistiramDormindo int32                         // cliente foi embora vendo todos dormindo
	desistiramCheio    int32                         // cliente não conseguiu entrar (fila cheia)
)

type Barbeiro struct { //construtor do tipo barbeiro
	id       int  // id do barbeiro
	dormindo bool // informação sobre se o mesmo esta dormindo
}

func (b *Barbeiro) cortar(cliente int) {
	mu.Lock() //bloqueia o mutex bloquando o barbeiro de tal id realizar ações
	fmt.Printf("Barbeiro %d cortando cliente %d\n", b.id, cliente)
	tempo := 2 + rand.Intn(2) // tempo do corte entre 2 -3 seg rand.Intn(2) gera ou 0 ou 1.
	time.Sleep(time.Duration(tempo) * time.Second)
	fmt.Printf("Barbeiro %d terminou cliente %d\n", b.id, cliente)
	mu.Unlock() // desbloqueio do barbeiro pelo mutex

	atomic.AddInt32(&atendidos, 1) // atomic faz com que o incremento dos atendidos nao quebre devido as diversas rotinas simultaneas
}

func (b *Barbeiro) trabalhar() {
	for {
		select { // divide em 2 casos
		case cliente := <-cadeiras: // caso 1 tem cliente na fila e o barbeiro acorda e corta
			b.dormindo = false
			b.cortar(cliente)
		default:
			if rand.Intn(100) < chanceSono { // caso 2 nao tem clientes entao o barbeiro pode ou não dormir
				b.dormindo = true
				tempo := 5 + rand.Intn(6)
				fmt.Printf("Barbeiro %d dormindo por %ds\n", b.id, tempo)
				time.Sleep(time.Duration(tempo) * time.Second) // fala qual barbeiro dormiu ou acordou e o tempo que tal dormiu.
				fmt.Printf("Barbeiro %d acordou apos %ds\n", b.id, tempo)
				b.dormindo = false
			} else {
				time.Sleep(500 * time.Millisecond) //reinicia o loop sem consumir muita cpu colocando uma mini espera
			}
		}
	}
}

func desistencias() {
	for {
		time.Sleep(1 * time.Second) // roda o loop infinitamente a cada 1 seg
		if len(cadeiras) == 0 {
			continue
		}
		if rand.Intn(chanceDesistir) == 0 { //chanceDesistir é 30 logo temos 1 chance em 30 de desistencia
			n := len(cadeiras)
			tmp := make([]int, 0, n) //cria a variavel tmp que sao slices de inteiros e que começa vazia e vai ate n
			for i := 0; i < n; i++ {
				tmp = append(tmp, <-cadeiras)
			}
			escolhido := rand.Intn(len(tmp)) // escolhe um cliente aleatorio para desistir
			fmt.Printf("Cliente %d desistiu da fila\n", tmp[escolhido])
			for i, id := range tmp {
				if i != escolhido {
					cadeiras <- id
				}
			}
			atomic.AddInt32(&desistiramFila, 1) // mais uma vez o atomic sendo usado para nao gerar conflito na contagem.
		}
	}
}

func todosDormindo() bool {
	for _, b := range barbeiros { // para cada barbeiro, se qualquer um barbeiro estiver acordado retorna false senao retorna True
		if !b.dormindo {
			return false
		}
	}
	return true
}

func mostrarContadores() {
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("\n------ ESTATÍSTICAS ------")
		fmt.Printf("Atendidos: %d\n", atomic.LoadInt32(&atendidos)) // quando varias routines podem alterar o mesmo valor o ideal é usar atomicidade na leitura dos dados.
		fmt.Printf("Desistiram da fila: %d\n", atomic.LoadInt32(&desistiramFila))
		fmt.Printf("Desistiram ao ver todos dormindo: %d\n", atomic.LoadInt32(&desistiramDormindo))
		fmt.Printf("Foram embora (fila cheia): %d\n", atomic.LoadInt32(&desistiramCheio))
		fmt.Println("--------------------------\n")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano()) //

	for i := 1; i <= numBarbeiros; i++ { // gera o numero de barbeiros indicado em numBarbeiros
		b := &Barbeiro{id: i}
		barbeiros = append(barbeiros, b)
		go b.trabalhar()
	}

	go desistencias()
	go mostrarContadores()

	clienteID := 1
	for {
		time.Sleep(2 * time.Second)
		fmt.Printf("Cliente %d chegou\n", clienteID)

		if todosDormindo() {
			if rand.Intn(2) == 0 { // chance de 50% de desistir ou nao , pois ou é gerado ou 0 ou 1
				fmt.Printf("Cliente %d viu todos dormindo e foi embora\n", clienteID)
				atomic.AddInt32(&desistiramDormindo, 1) //contagem de forma atomica
				clienteID++                             //passa para o proximo cliente
				continue
			}
		}

		select {
		case cadeiras <- clienteID: // tenta enviar cliente para cadeiras checando assim se ha vagas.
			fmt.Printf("Cliente %d sentou (espera: %d/%d)\n", clienteID, len(cadeiras), numCadeiras)
		default:
			fmt.Printf("Cliente %d foi embora (fila cheia)\n", clienteID)
			atomic.AddInt32(&desistiramCheio, 1)
		}
		clienteID++
	}
}
