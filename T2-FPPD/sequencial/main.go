package main

import (
	"fmt"
	"math/rand"
	"time"
)

const N = 3000

// Gera duas matrizes aleatórias A e B
func gerarMatrizes(n int) ([][]float64, [][]float64) {

	A := make([][]float64, n)
	B := make([][]float64, n)

	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
		B[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			A[i][j] = rand.Float64() * 10
			B[i][j] = rand.Float64() * 10
		}
	}

	return A, B
}

// Multiplica duas matrizes usando o algoritmo ingênuo (triplo loop)
func multiplicarMatrizes(A, B [][]float64, n int) [][]float64 {

	C := make([][]float64, n)

	for i := 0; i < n; i++ {
		C[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {

			soma := 0.0

			for k := 0; k < n; k++ {
				soma += A[i][k] * B[k][j]
			}

			C[i][j] = soma
		}
	}

	return C
}

func main() {

	// Seed fixa para gerar sempre as mesmas matrizes
	rand.Seed(42)

	// Gera as matrizes A e B
	A, B := gerarMatrizes(N)

	// Inicia a medição do tempo
	inicio := time.Now()

	// Multiplica as matrizes
	C := multiplicarMatrizes(A, B, N)

	// Finaliza a medição
	tempo := time.Since(inicio)

	fmt.Printf("Tempo de execução: %v\n", tempo)

	// Valores de verificação
	fmt.Println("\nValores de verificação:")

	fmt.Printf("C[0][0] = %.2f\n", C[0][0])
	fmt.Printf("C[0][N-1] = %.2f\n", C[0][N-1])
	fmt.Printf("C[N-1][0] = %.2f\n", C[N-1][0])
	fmt.Printf("C[N-1][N-1] = %.2f\n", C[N-1][N-1])
}