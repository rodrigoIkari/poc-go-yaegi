package main

import (
	"fmt"
	"go/build"
	"math"
	"os"
	"poc-go-yaegi/functions"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// TestScriptExecution_HappyPath testa o cenário feliz de execução do script
func TestScriptExecution_HappyPath(t *testing.T) {
	// Configurar o interpretador Yaegi
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	// Registrar bibliotecas padrão
	i.Use(interp.Exports{
		"math/math":       stdlib.Symbols["math/math"],
		"fmt/fmt":         stdlib.Symbols["fmt/fmt"],
		"time/time":       stdlib.Symbols["time/time"],
		"strconv/strconv": stdlib.Symbols["strconv/strconv"],
	})

	// Registrar tipos customizados
	i.Use(interp.Exports{
		"poc-go-yaegi/functions/functions": map[string]reflect.Value{
			"Operation":   reflect.ValueOf((*functions.Operation)(nil)),
			"Installment": reflect.ValueOf((*functions.Installment)(nil)),
			"Component":   reflect.ValueOf((*functions.Component)(nil)),
		},
	})

	i.ImportUsed()

	// Carregar o script
	scriptFile, err := os.ReadFile("script/script.go")
	if err != nil {
		t.Fatalf("Erro ao ler arquivo de script: %v", err)
	}

	scriptCode := string(scriptFile)

	// Compilar o script
	_, err = i.Eval(scriptCode)
	if err != nil {
		t.Fatalf("Erro ao compilar script: %v", err)
	}

	// Obter a função do script
	script, err := i.Eval("script.CalculateOperationStep")
	if err != nil {
		t.Fatalf("Erro ao obter função CalculateOperationStep: %v", err)
	}

	// Type assertion
	calculateOperationStep, ok := script.Interface().(func(map[string]string, functions.Operation) (*functions.Operation, error))
	if !ok {
		t.Fatal("Função CalculateOperationStep tem assinatura incorreta")
	}

	// Preparar dados de entrada
	params := map[string]string{
		"qtde_parcelas": "12",
		"taxa_juros":    "0.02",
	}

	inputOperation := functions.Operation{
		Amount: 1000.00,
	}

	// Executar a função
	result, err := calculateOperationStep(params, inputOperation)
	if err != nil {
		t.Fatalf("Erro ao executar script: %v", err)
	}

	// Validações do resultado
	if result == nil {
		t.Fatal("Resultado não deveria ser nil")
	}

	if result.Amount != 1000.00 {
		t.Errorf("Amount incorreto: esperado 1000.00, obtido %.2f", result.Amount)
	}

	if result.InstallmentsQty != 12 {
		t.Errorf("InstallmentsQty incorreto: esperado 12, obtido %d", result.InstallmentsQty)
	}

	// Validar que o total é maior que o principal (devido aos juros)
	if result.TotalAmount <= result.Amount {
		t.Errorf("TotalAmount (%.2f) deveria ser maior que Amount (%.2f)", result.TotalAmount, result.Amount)
	}

	// Validar quantidade de parcelas
	if len(result.Installments) != 12 {
		t.Errorf("Número de parcelas incorreto: esperado 12, obtido %d", len(result.Installments))
	}

	// Validar primeira parcela
	if len(result.Installments) > 0 {
		firstInstallment := result.Installments[0]

		if firstInstallment.Number != 1 {
			t.Errorf("Número da primeira parcela incorreto: esperado 1, obtido %d", firstInstallment.Number)
		}

		if firstInstallment.Amount <= 0 {
			t.Errorf("Valor da parcela deveria ser positivo, obtido %.2f", firstInstallment.Amount)
		}

		// Validar componentes da parcela
		if len(firstInstallment.Components) != 2 {
			t.Errorf("Primeira parcela deveria ter 2 componentes, obtido %d", len(firstInstallment.Components))
		}

		// Verificar se tem componente Principal
		hasPrincipal := false
		hasJuros := false
		for _, comp := range firstInstallment.Components {
			if comp.Name == "Principal" {
				hasPrincipal = true
			}
			if comp.Name == "Juros" {
				hasJuros = true
			}
		}

		if !hasPrincipal {
			t.Error("Primeira parcela deveria ter componente 'Principal'")
		}

		if !hasJuros {
			t.Error("Primeira parcela deveria ter componente 'Juros'")
		}
	}

	// Validar que a soma das parcelas é aproximadamente igual ao total
	// (pode haver pequenas diferenças por arredondamento)
	somaaParcelas := 0.0
	for _, installment := range result.Installments {
		somaaParcelas += installment.Amount
	}

	diferenca := result.TotalAmount - somaaParcelas
	if diferenca > 0.01 || diferenca < -0.01 {
		t.Errorf("Soma das parcelas (%.2f) diverge do total (%.2f) em %.2f",
			somaaParcelas, result.TotalAmount, diferenca)
	}

	// Validar cálculo PRICE esperado
	// Para 1000 reais, 12 parcelas, 2% ao mês
	// Valor esperado por parcela: ~94.56
	expectedInstallment := 94.56
	tolerance := 0.10 // tolerância de 10 centavos

	if result.Installments[0].Amount < expectedInstallment-tolerance ||
	   result.Installments[0].Amount > expectedInstallment+tolerance {
		t.Errorf("Valor da parcela fora do esperado: esperado ~%.2f, obtido %.2f",
			expectedInstallment, result.Installments[0].Amount)
	}
}

// TestScriptExecution_InvalidParams testa execução com parâmetros inválidos
func TestScriptExecution_InvalidParams(t *testing.T) {
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	i.Use(interp.Exports{
		"math/math":       stdlib.Symbols["math/math"],
		"fmt/fmt":         stdlib.Symbols["fmt/fmt"],
		"time/time":       stdlib.Symbols["time/time"],
		"strconv/strconv": stdlib.Symbols["strconv/strconv"],
	})

	i.Use(interp.Exports{
		"poc-go-yaegi/functions/functions": map[string]reflect.Value{
			"Operation":   reflect.ValueOf((*functions.Operation)(nil)),
			"Installment": reflect.ValueOf((*functions.Installment)(nil)),
			"Component":   reflect.ValueOf((*functions.Component)(nil)),
		},
	})

	i.ImportUsed()

	scriptFile, err := os.ReadFile("script/script.go")
	if err != nil {
		t.Fatalf("Erro ao ler arquivo de script: %v", err)
	}

	_, err = i.Eval(string(scriptFile))
	if err != nil {
		t.Fatalf("Erro ao compilar script: %v", err)
	}

	script, err := i.Eval("script.CalculateOperationStep")
	if err != nil {
		t.Fatalf("Erro ao obter função: %v", err)
	}

	calculateOperationStep, ok := script.Interface().(func(map[string]string, functions.Operation) (*functions.Operation, error))
	if !ok {
		t.Fatal("Type assertion falhou")
	}

	// Teste com parâmetros faltando
	t.Run("MissingQtdeParcelas", func(t *testing.T) {
		params := map[string]string{
			"taxa_juros": "0.02",
		}
		inputOperation := functions.Operation{Amount: 1000.00}

		_, err := calculateOperationStep(params, inputOperation)
		if err == nil {
			t.Error("Deveria retornar erro quando qtde_parcelas está faltando")
		}
	})

	// Teste com parâmetros inválidos
	t.Run("InvalidQtdeParcelas", func(t *testing.T) {
		params := map[string]string{
			"qtde_parcelas": "abc",
			"taxa_juros":    "0.02",
		}
		inputOperation := functions.Operation{Amount: 1000.00}

		_, err := calculateOperationStep(params, inputOperation)
		if err == nil {
			t.Error("Deveria retornar erro quando qtde_parcelas é inválida")
		}
	})
}

// TestScriptWithUnregisteredLibrary testa que o script falha ao usar biblioteca não registrada
func TestScriptWithUnregisteredLibrary(t *testing.T) {
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	// Registrar apenas bibliotecas básicas, SEM math/rand
	i.Use(interp.Exports{
		"fmt/fmt":         stdlib.Symbols["fmt/fmt"],
		"time/time":       stdlib.Symbols["time/time"],
		"strconv/strconv": stdlib.Symbols["strconv/strconv"],
	})

	i.Use(interp.Exports{
		"poc-go-yaegi/functions/functions": map[string]reflect.Value{
			"Operation":   reflect.ValueOf((*functions.Operation)(nil)),
			"Installment": reflect.ValueOf((*functions.Installment)(nil)),
			"Component":   reflect.ValueOf((*functions.Component)(nil)),
		},
	})

	i.ImportUsed()

	// Carregar script que usa math/rand (não registrada)
	scriptFile, err := os.ReadFile("script/script_with_random.go")
	if err != nil {
		t.Fatalf("Erro ao ler script de teste: %v", err)
	}

	// Este script usa math/rand que NÃO foi registrado
	_, err = i.Eval(string(scriptFile))

	// Esperamos que o erro ocorra
	if err == nil {
		t.Fatal("Esperava erro ao usar biblioteca não registrada (math/rand), mas não ocorreu erro")
	}

	// Verificar que a mensagem de erro menciona o problema
	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("Mensagem de erro está vazia")
	}

	t.Logf("Erro esperado recebido: %v", err)

	// Verificar se a mensagem menciona a biblioteca não encontrada
	// O Yaegi tipicamente retorna erro como "unable to find source related to"
	if len(errorMsg) > 0 {
		// Apenas log do erro, já validamos que ele existe
		t.Logf("Biblioteca não registrada causou erro como esperado: %s", errorMsg)
	}
}

// TestPerformance_1000Executions testa performance com 1000 execuções
// Mede tempo mínimo, médio, P95 e máximo de execução
func TestPerformance_1000Executions(t *testing.T) {
	// Configurar o interpretador Yaegi (setup único)
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	i.Use(interp.Exports{
		"math/math":       stdlib.Symbols["math/math"],
		"fmt/fmt":         stdlib.Symbols["fmt/fmt"],
		"time/time":       stdlib.Symbols["time/time"],
		"strconv/strconv": stdlib.Symbols["strconv/strconv"],
	})

	i.Use(interp.Exports{
		"poc-go-yaegi/functions/functions": map[string]reflect.Value{
			"Operation":   reflect.ValueOf((*functions.Operation)(nil)),
			"Installment": reflect.ValueOf((*functions.Installment)(nil)),
			"Component":   reflect.ValueOf((*functions.Component)(nil)),
		},
	})

	i.ImportUsed()

	// Carregar e compilar o script silencioso (sem prints) para performance
	scriptFile, err := os.ReadFile("script/script_silent.go")
	if err != nil {
		t.Fatalf("Erro ao ler arquivo de script: %v", err)
	}

	_, err = i.Eval(string(scriptFile))
	if err != nil {
		t.Fatalf("Erro ao compilar script: %v", err)
	}

	script, err := i.Eval("script.CalculateOperationStepSilent")
	if err != nil {
		t.Fatalf("Erro ao obter função: %v", err)
	}

	calculateOperationStep, ok := script.Interface().(func(map[string]string, functions.Operation) (*functions.Operation, error))
	if !ok {
		t.Fatal("Type assertion falhou")
	}

	// Preparar dados de entrada (reutilizados em todas as execuções)
	params := map[string]string{
		"qtde_parcelas": "12",
		"taxa_juros":    "0.02",
	}

	inputOperation := functions.Operation{
		Amount: 1000.00,
	}

	// Array para armazenar os tempos de execução
	const numExecutions = 1000
	executionTimes := make([]time.Duration, numExecutions)

	fmt.Printf("\n=== Teste de Performance - %d Execuções ===\n", numExecutions)
	fmt.Println("Executando...")

	// Executar 1000 vezes e medir o tempo de cada execução
	for i := 0; i < numExecutions; i++ {
		startTime := time.Now()

		_, err := calculateOperationStep(params, inputOperation)
		if err != nil {
			t.Fatalf("Erro na execução %d: %v", i+1, err)
		}

		executionTimes[i] = time.Since(startTime)

		// Mostrar progresso a cada 100 execuções
		if (i+1)%100 == 0 {
			fmt.Printf("  Progresso: %d/%d execuções\n", i+1, numExecutions)
		}
	}

	// Ordenar os tempos para calcular percentis
	sortedTimes := make([]time.Duration, numExecutions)
	copy(sortedTimes, executionTimes)
	sort.Slice(sortedTimes, func(i, j int) bool {
		return sortedTimes[i] < sortedTimes[j]
	})

	// Calcular estatísticas
	minTime := sortedTimes[0]
	maxTime := sortedTimes[numExecutions-1]

	// Calcular média
	var totalTime time.Duration
	for _, t := range executionTimes {
		totalTime += t
	}
	avgTime := totalTime / numExecutions

	// Calcular P95 (percentil 95)
	p95Index := int(math.Ceil(float64(numExecutions) * 0.95)) - 1
	if p95Index >= numExecutions {
		p95Index = numExecutions - 1
	}
	p95Time := sortedTimes[p95Index]

	// Calcular P50 (mediana)
	p50Index := numExecutions / 2
	p50Time := sortedTimes[p50Index]

	// Calcular desvio padrão
	var sumSquaredDiff float64
	for _, t := range executionTimes {
		diff := float64(t - avgTime)
		sumSquaredDiff += diff * diff
	}
	stdDev := time.Duration(math.Sqrt(sumSquaredDiff / float64(numExecutions)))

	// Exibir resultados
	fmt.Println("\n=== Resultados do Teste de Performance ===")
	fmt.Printf("Número de Execuções: %d\n\n", numExecutions)
	fmt.Printf("Tempo Mínimo:   %v\n", minTime)
	fmt.Printf("Tempo Médio:    %v\n", avgTime)
	fmt.Printf("Tempo Mediana:  %v (P50)\n", p50Time)
	fmt.Printf("Tempo P95:      %v\n", p95Time)
	fmt.Printf("Tempo Máximo:   %v\n", maxTime)
	fmt.Printf("Desvio Padrão:  %v\n\n", stdDev)

	// Conversão para microsegundos para melhor visualização
	fmt.Println("=== Em Microsegundos (µs) ===")
	fmt.Printf("Mínimo:   %d µs\n", minTime.Microseconds())
	fmt.Printf("Médio:    %d µs\n", avgTime.Microseconds())
	fmt.Printf("Mediana:  %d µs (P50)\n", p50Time.Microseconds())
	fmt.Printf("P95:      %d µs\n", p95Time.Microseconds())
	fmt.Printf("Máximo:   %d µs\n", maxTime.Microseconds())
	fmt.Printf("Std Dev:  %d µs\n\n", stdDev.Microseconds())

	// Calcular throughput
	totalTimeSeconds := totalTime.Seconds()
	throughput := float64(numExecutions) / totalTimeSeconds
	fmt.Printf("Throughput: %.2f execuções/segundo\n", throughput)
	fmt.Printf("Tempo Total: %v\n\n", totalTime)

	// Logging para o test output
	t.Logf("Performance Stats - %d execuções:", numExecutions)
	t.Logf("  Min: %v | Avg: %v | P50: %v | P95: %v | Max: %v",
		minTime, avgTime, p50Time, p95Time, maxTime)
	t.Logf("  Throughput: %.2f exec/s", throughput)

	// Validações para garantir que o desempenho está aceitável
	// Ajuste estes valores conforme necessário
	if avgTime > 10*time.Millisecond {
		t.Logf("AVISO: Tempo médio (%.2fms) está acima do esperado", avgTime.Seconds()*1000)
	}

	if p95Time > 20*time.Millisecond {
		t.Logf("AVISO: P95 (%.2fms) está acima do esperado", p95Time.Seconds()*1000)
	}
}
