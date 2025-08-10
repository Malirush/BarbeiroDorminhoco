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
	numClientes    = 10 // número fixo de clientes
	chanceSono     = 30
	chanceDesistir = 30
)

var ( // variaveis
	cadeiras           = make(chan int, numCadeiras)     // cria o canal para as cadeiras
	mu                 sync.Mutex                        // mutex para bloquear se necessário
	barbeiros          []*Barbeiro                       // lista de barbeiros
	atendidos          int32                             // total atendidos
	desistiramFila     int32                             // clientes que desistiram depois de entrar
	desistiramDormindo int32                             // clientes que desistiram ao ver todos dormindo
	desistiramCheio    int32                             // clientes que não conseguiram entrar, fila cheia
	dia                int32                         = 1 // contador de dias
	cont               int32                             // contador de clientes processados
	clientesAvaliados  = make(map[int]bool)              //map para verificar se cliente ja foi avaliado

	avaliacoes []Avaliacao // Slice para armazenar todas as avaliações
)

type Barbeiro struct { // tipo barbeiro
	id       int  // id do barbeiro
	dormindo bool // status de se o barbeiro está dormindo
}

// Estrutura para armazenar a avaliação de cada cliente (apenas a nota)
type Avaliacao struct {
	clienteID int
	nota      int
}

func (b *Barbeiro) cortar(cliente int) {
	mu.Lock() // bloqueia o mutex para evitar competição
	fmt.Printf("Barbeiro %d cortando cliente %d\n", b.id, cliente)
	tempo := 2 + rand.Intn(2) // tempo do corte entre 2 e 3 segundos
	time.Sleep(time.Duration(tempo) * time.Second)
	fmt.Printf("Barbeiro %d terminou cliente %d\n", b.id, cliente)
	avaliarCliente(cliente)
	atomic.AddInt32(&cont, 1)
	mu.Unlock() // libera o mutex

	atomic.AddInt32(&atendidos, 1) // incrementa o número de atendimentos
}

func (b *Barbeiro) trabalhar() {
	for {
		select { // se o cliente estiver na fila, corta o cabelo
		case cliente := <-cadeiras:
			b.dormindo = false
			b.cortar(cliente)
		default: // se não tiver clientes, o barbeiro pode dormir
			if rand.Intn(100) < chanceSono {
				b.dormindo = true
				tempo := 5 + rand.Intn(6)
				fmt.Printf("Barbeiro %d dormindo por %ds\n", b.id, tempo)
				time.Sleep(time.Duration(tempo) * time.Second)
				fmt.Printf("Barbeiro %d acordou após %ds\n", b.id, tempo)
				b.dormindo = false
			} else {
				time.Sleep(500 * time.Millisecond) // aguarda sem consumir CPU
			}
		}
	}
}

func desistencias() {
	for {
		time.Sleep(1 * time.Second)
		if len(cadeiras) == 0 {
			continue
		}
		if rand.Intn(chanceDesistir) == 0 { // chance de desistência
			n := len(cadeiras)
			tmp := make([]int, 0, n)
			for i := 0; i < n; i++ {
				tmp = append(tmp, <-cadeiras)
			}
			escolhido := rand.Intn(len(tmp))
			fmt.Printf("Cliente %d desistiu da fila\n", tmp[escolhido])
			atomic.AddInt32(&cont, 1)
			for i, id := range tmp {
				if i != escolhido {
					cadeiras <- id
				}
			}
			atomic.AddInt32(&desistiramFila, 1)
		}
	}
}

func todosDormindo() bool {
	for _, b := range barbeiros {
		if !b.dormindo {
			return false
		}
	}
	return true
}

func cliente(id int) {
	if todosDormindo() {
		if rand.Intn(2) == 0 {
			fmt.Printf("Cliente %d viu todos dormindo e foi embora\n", id)
			atomic.AddInt32(&cont, 1)
			atomic.AddInt32(&desistiramDormindo, 1)
			return
		}
	}

	select {
	case cadeiras <- id:
		fmt.Printf("Cliente %d sentou (espera: %d/%d)\n", id, len(cadeiras), numCadeiras)
	default:
		fmt.Printf("Cliente %d foi embora (fila cheia)\n", id)
		atomic.AddInt32(&cont, 1)
		atomic.AddInt32(&desistiramCheio, 1)
	}
}

// Função para gerar uma avaliação aleatória para o cliente (somente a nota)
func avaliarCliente(clienteID int) {
	if clientesAvaliados[clienteID] { // Verifica se o cliente já foi avaliado
		return // Não faz nada se já foi avaliado
	}

	// Gera uma avaliação aleatória (somente a nota)
	nota := rand.Intn(5) + 1 // Avaliação aleatória de 1 a 5

	// Cria a avaliação com a nota
	avaliacao := Avaliacao{clienteID: clienteID, nota: nota}
	avaliacoes = append(avaliacoes, avaliacao)

	// Marca o cliente como avaliado
	clientesAvaliados[clienteID] = true

	// Exibe a avaliação (apenas a nota)
	fmt.Printf("Cliente %d avaliou: %d\n", clienteID, nota)
}

// Goroutine para realizar avaliações dos clientes (a cada 3 segundos)
func avaliarClientes() {
	for {
		time.Sleep(3 * time.Second)

		// Exibe a média das avaliações periodicamente (por exemplo, a cada 10 segundos)
		if len(avaliacoes) > 0 {
			var somaNotas int
			for _, avaliacao := range avaliacoes {
				somaNotas += avaliacao.nota
			}
			media := float64(somaNotas) / float64(len(avaliacoes))
			fmt.Printf("Média das avaliações: %.2f\n", media)
		}
	}
}

func mostrarContadores() {

	fmt.Println("\n------ ESTATÍSTICAS ------")
	fmt.Printf("Dia: %d\n", dia)
	fmt.Printf("Atendidos: %d\n", atomic.LoadInt32(&atendidos))
	fmt.Printf("Desistiram da fila: %d\n", atomic.LoadInt32(&desistiramFila))
	fmt.Printf("Desistiram ao ver todos dormindo: %d\n", atomic.LoadInt32(&desistiramDormindo))
	fmt.Printf("Foram embora (fila cheia): %d\n", atomic.LoadInt32(&desistiramCheio))
	fmt.Println("--------------------------\n")

}

func main() {
	rand.Seed(time.Now().UnixNano()) //

	for i := 1; i <= numBarbeiros; i++ {
		b := &Barbeiro{id: i}
		barbeiros = append(barbeiros, b)
		go b.trabalhar()
	}

	go desistencias()
	go avaliarClientes()

	// Goroutines para os 10 clientes fixos
	for i := 1; i <= numClientes; i++ {
		go cliente(i)
		time.Sleep(2 * time.Second) // Intervalo de chegada de clientes
	}

	// Acompanhamento do ciclo de dias
	for {
		time.Sleep(2 * time.Second)
		// Quando todos os clientes forem atendidos, incrementa o dia e reinicia os clientes
		if atomic.LoadInt32(&cont) == numClientes {
			go mostrarContadores()
			dia++
			fmt.Printf("\n--- DIA %d INICIADO ---\n", dia)
			atomic.StoreInt32(&cont, 0)
			// Reinicia as avaliações dos clientes
			for id := range clientesAvaliados {
				clientesAvaliados[id] = false // Marca todos os clientes como não avaliados
			}
			// Reinicia as goroutines dos clientes
			for i := 1; i <= numClientes; i++ {
				go cliente(i)
				time.Sleep(2 * time.Second)
			}
		}
	}
}
