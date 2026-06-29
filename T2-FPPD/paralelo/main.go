// Versão paralela (MPI) da multiplicação de matrizes C = A x B
// Modelo: Mestre-Escravo (Master-Worker)
//   - O processo de rank 0 (mestre) gera as matrizes A e B, distribui blocos
//     de linhas de A (mais a matriz B completa) para os escravos, e coleta os
//     blocos de C calculados por cada um.
//   - Os processos escravos (rank > 0) recebem seu bloco de linhas de A e a
//     matriz B completa, calculam o bloco correspondente de C, e devolvem o
//     resultado ao mestre.
//   - Decomposição de dados: por LINHAS de A. Cada processo p recebe um
//     intervalo contíguo de linhas [inicio, fim). Se N não for divisível
//     igualmente pelo número de processos, as linhas restantes são
//     distribuídas entre os primeiros processos (um a mais cada).
//
// FPPD - T2 Processamento Paralelo - PUCRS 2026/1
//
// Execução: mpirun -np <P> ./paralelo -n 3000 -seed 42
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	mpi "github.com/mnlphlp/gompi"
)

// gerarMatriz cria uma matriz N x N (row-major) com valores pseudoaleatórios,
// idêntica à função usada na versão sequencial, para garantir reprodutibilidade.
func gerarMatriz(n int, seed int64) []float64 {
	r := rand.New(rand.NewSource(seed))
	m := make([]float64, n*n)
	for i := range m {
		m[i] = r.Float64()
	}
	return m
}

// checksum soma todos os elementos de C (mesma função da versão sequencial).
func checksum(c []float64) float64 {
	var soma float64
	for _, v := range c {
		soma += v
	}
	return soma
}

// calcularLinhas multiplica um bloco de linhas de A (linhaIni:linhaFim) pela
// matriz B completa, retornando o bloco correspondente de C.
func calcularLinhas(a []float64, b []float64, n int, linhaIni, linhaFim int) []float64 {
	numLinhas := linhaFim - linhaIni
	c := make([]float64, numLinhas*n)
	for li := 0; li < numLinhas; li++ {
		i := linhaIni + li
		for k := 0; k < n; k++ {
			aik := a[i*n+k]
			if aik == 0 {
				continue
			}
			for j := 0; j < n; j++ {
				c[li*n+j] += aik * b[k*n+j]
			}
		}
	}
	return c
}

// distribuirLinhas calcula, para um dado número de processos, quantas linhas
// (e quais intervalos) cabem a cada processo. As linhas restantes da divisão
// inteira são distribuídas uma a uma entre os primeiros processos.
func distribuirLinhas(n, numProcessos int) (inicios, fins []int) {
	inicios = make([]int, numProcessos)
	fins = make([]int, numProcessos)
	base := n / numProcessos
	resto := n % numProcessos
	atual := 0
	for p := 0; p < numProcessos; p++ {
		linhas := base
		if p < resto {
			linhas++ // distribui o resto entre os primeiros processos
		}
		inicios[p] = atual
		atual += linhas
		fins[p] = atual
	}
	return
}

func main() {
	n := flag.Int("n", 3000, "dimensão das matrizes (N x N)")
	seed := flag.Int64("seed", 42, "seed para geração das matrizes")
	flag.Parse()

	mpi.Init()
	defer mpi.Finalize()

	comm := mpi.NewComm(true) // true: panic em caso de erro de comunicação
	rank := comm.GetRank()
	numProcessos := comm.GetSize()

	N := *n

	const (
		tagLinhasA = 1 // mestre -> escravo: bloco de linhas de A
		tagB       = 2 // mestre -> escravo: matriz B completa
		tagInicio  = 3 // mestre -> escravo: índice da primeira linha do bloco
		tagResultC = 4 // escravo -> mestre: bloco de linhas de C calculado
	)

	inicios, fins := distribuirLinhas(N, numProcessos)

	if rank == 0 {
		// ============ PROCESSO MESTRE ============
		fmt.Printf("Versão PARALELA (MPI, Mestre-Escravo) - N=%d, processos=%d, seed=%d\n", N, numProcessos, *seed)

		// Geração das matrizes (fora da medição de tempo).
		a := gerarMatriz(N, *seed)
		b := gerarMatriz(N, *seed+1)

		// --- Início da medição: distribuição + cálculo + coleta ---
		inicio := time.Now()

		// Distribui para os escravos (ranks 1..P-1) seus blocos de linhas de A
		// e a matriz B completa.
		for p := 1; p < numProcessos; p++ {
			li, lf := inicios[p], fins[p]
			blocoA := a[li*N : lf*N]
			comm.Send([]int{li, lf}, p, tagInicio)
			comm.Send(blocoA, p, tagLinhasA)
			comm.Send(b, p, tagB)
		}

		// O próprio mestre calcula seu bloco de linhas (rank 0).
		c := make([]float64, N*N)
		liMestre, lfMestre := inicios[0], fins[0]
		blocoMestre := calcularLinhas(a, b, N, liMestre, lfMestre)
		copy(c[liMestre*N:lfMestre*N], blocoMestre)

		// Coleta os resultados de cada escravo.
		for p := 1; p < numProcessos; p++ {
			li, lf := inicios[p], fins[p]
			var blocoC []float64
			comm.Recv(&blocoC, p, tagResultC)
			copy(c[li*N:lf*N], blocoC)
		}

		duracao := time.Since(inicio)
		// --- Fim da medição ---

		fmt.Printf("Tempo de execução (paralelo, %d processos): %v\n", numProcessos, duracao)
		fmt.Printf("Tempo em segundos: %.6f\n", duracao.Seconds())

		fmt.Printf("C[0][0]       = %.6f\n", c[0])
		fmt.Printf("C[0][N-1]     = %.6f\n", c[N-1])
		fmt.Printf("C[N-1][0]     = %.6f\n", c[(N-1)*N])
		fmt.Printf("C[N-1][N-1]   = %.6f\n", c[(N-1)*N+(N-1)])
		fmt.Printf("Checksum(C)   = %.6f\n", checksum(c))

	} else {
		// ============ PROCESSO ESCRAVO ============
		var intervalo []int
		comm.Recv(&intervalo, 0, tagInicio)
		li, lf := intervalo[0], intervalo[1]

		var blocoA []float64
		comm.Recv(&blocoA, 0, tagLinhasA)

		var b []float64
		comm.Recv(&b, 0, tagB)

		// Recalcula o bloco local usando índices relativos (0..numLinhas),
		// já que blocoA contém apenas as linhas recebidas.
		numLinhas := lf - li
		c := make([]float64, numLinhas*N)
		for li2 := 0; li2 < numLinhas; li2++ {
			for k := 0; k < N; k++ {
				aik := blocoA[li2*N+k]
				if aik == 0 {
					continue
				}
				for j := 0; j < N; j++ {
					c[li2*N+j] += aik * b[k*N+j]
				}
			}
		}

		comm.Send(c, 0, tagResultC)
	}
}