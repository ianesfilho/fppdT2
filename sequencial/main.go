package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

// gerarMatriz cria uma matriz N x N (row-major)
// utilizando uma seed fixa para garantir reprodutibilidade.
func gerarMatriz(n int, seed int64) []float64 {

	r := rand.New(rand.NewSource(seed))

	matriz := make([]float64, n*n)

	for i := range matriz {
		matriz[i] = r.Float64()
	}

	return matriz
}

// checksum soma todos os elementos da matriz.
func checksum(c []float64) float64 {

	var soma float64

	for _, valor := range c {
		soma += valor
	}

	return soma
}

// multiplicarMatrizes realiza a multiplicação ingênua
// C = A x B.
func multiplicarMatrizes(a, b []float64, n int) []float64 {

	c := make([]float64, n*n)

	for i := 0; i < n; i++ {

		for k := 0; k < n; k++ {

			aik := a[i*n+k]

			if aik == 0 {
				continue
			}

			for j := 0; j < n; j++ {

				c[i*n+j] += aik * b[k*n+j]

			}
		}

	}

	return c
}

func main() {

	n := flag.Int("n", 3000, "Dimensão da matriz (N x N)")
	seed := flag.Int64("seed", 42, "Seed para geração das matrizes")

	flag.Parse()

	N := *n

	fmt.Printf("Versão SEQUENCIAL - N=%d, seed=%d\n", N, *seed)

	// Geração das matrizes (fora da medição de tempo)
	a := gerarMatriz(N, *seed)
	b := gerarMatriz(N, *seed+1)

	fmt.Println("Matrizes geradas.")
	fmt.Println("Iniciando multiplicação...")

	// Início da medição
	inicio := time.Now()

	c := multiplicarMatrizes(a, b, N)

	duracao := time.Since(inicio)

	// Fim da medição

	fmt.Printf("\nTempo de execução (sequencial): %v\n", duracao)
	fmt.Printf("Tempo em segundos: %.6f\n\n", duracao.Seconds())

	fmt.Printf("C[0][0]     = %.6f\n", c[0])
	fmt.Printf("C[0][N-1]   = %.6f\n", c[N-1])
	fmt.Printf("C[N-1][0]   = %.6f\n", c[(N-1)*N])
	fmt.Printf("C[N-1][N-1] = %.6f\n", c[(N-1)*N+(N-1)])

	fmt.Printf("Checksum(C) = %.6f\n", checksum(c))
}