package main

import (
	"fmt"
	"go/build"
	"math"
	"math/rand"
	"os"
	"poc-go-yaegi/functions"
	"poc-go-yaegi/service"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// TestPerformance_ServiceVsYaegi_50000Executions compara performance do service nativo vs Yaegi
// com 50000 execuções e parâmetros variados
func TestPerformance_ServiceVsYaegi_50000Executions(t *testing.T) {
	const numExecutions = 50000

	// Seed para geração de valores aleatórios (para reprodutibilidade)
	rand.Seed(time.Now().UnixNano())

	fmt.Printf("\n=== Teste de Performance Comparativo - %d Execuções ===\n", numExecutions)
	fmt.Println("Comparando: Service Nativo vs Yaegi Script")
	fmt.Println("Variando: Valor do empréstimo (1000-100000), Taxa de juros (1%-5%), Parcelas (4-18)")
	fmt.Println()

	// ============================================
	// 1. TESTE COM SERVICE NATIVO
	// ============================================
	fmt.Println("=== Executando Service Nativo ===")

	nativeCalculator := service.NewPriceCalculator()
	nativeExecutionTimes := make([]time.Duration, numExecutions)

	// Gerar conjunto de dados de teste
	testData := generateTestData(numExecutions)

	fmt.Println("Executando...")
	for i := 0; i < numExecutions; i++ {
		params := testData[i].params
		inputOp := testData[i].operation

		startTime := time.Now()
		_, err := nativeCalculator.CalculateOperationStep(params, inputOp)
		if err != nil {
			t.Fatalf("Erro na execução nativa %d: %v", i+1, err)
		}
		nativeExecutionTimes[i] = time.Since(startTime)

		// Mostrar progresso a cada 5000 execuções
		if (i+1)%5000 == 0 {
			fmt.Printf("  Progresso: %d/%d execuções\n", i+1, numExecutions)
		}
	}

	nativeStats := calculateStats(nativeExecutionTimes)

	// ============================================
	// 2. TESTE COM YAEGI
	// ============================================
	fmt.Println("\n=== Executando Yaegi Script ===")

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

	// Carregar e compilar o script silencioso
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

	yaegiExecutionTimes := make([]time.Duration, numExecutions)

	fmt.Println("Executando...")
	for i := 0; i < numExecutions; i++ {
		params := testData[i].params
		inputOp := testData[i].operation

		startTime := time.Now()
		_, err := calculateOperationStep(params, inputOp)
		if err != nil {
			t.Fatalf("Erro na execução Yaegi %d: %v", i+1, err)
		}
		yaegiExecutionTimes[i] = time.Since(startTime)

		// Mostrar progresso a cada 5000 execuções
		if (i+1)%5000 == 0 {
			fmt.Printf("  Progresso: %d/%d execuções\n", i+1, numExecutions)
		}
	}

	yaegiStats := calculateStats(yaegiExecutionTimes)

	// ============================================
	// 3. EXIBIR RESULTADOS COMPARATIVOS
	// ============================================
	fmt.Println("\n======================================================================")
	fmt.Println("RESULTADOS DO TESTE DE PERFORMANCE")
	fmt.Println("======================================================================")
	fmt.Printf("Número de Execuções: %d\n", numExecutions)
	fmt.Printf("Parâmetros: Empréstimo (1000-100000), Juros (1%%-5%%), Parcelas (4-18)\n\n")

	// Tabela comparativa
	fmt.Println("┌─────────────────────┬──────────────────┬──────────────────┬──────────────────┐")
	fmt.Println("│ Métrica             │ Service Nativo   │ Yaegi Script     │ Diferença        │")
	fmt.Println("├─────────────────────┼──────────────────┼──────────────────┼──────────────────┤")

	printComparisonRow("Tempo Mínimo", nativeStats.Min, yaegiStats.Min)
	printComparisonRow("Tempo Médio", nativeStats.Avg, yaegiStats.Avg)
	printComparisonRow("Tempo Mediana (P50)", nativeStats.P50, yaegiStats.P50)
	printComparisonRow("Tempo P95", nativeStats.P95, yaegiStats.P95)
	printComparisonRow("Tempo Máximo", nativeStats.Max, yaegiStats.Max)
	printComparisonRow("Desvio Padrão", nativeStats.StdDev, yaegiStats.StdDev)

	fmt.Println("└─────────────────────┴──────────────────┴──────────────────┴──────────────────┘")

	// Throughput
	fmt.Println("\n┌─────────────────────┬──────────────────┬──────────────────┐")
	fmt.Println("│ Throughput          │ Service Nativo   │ Yaegi Script     │")
	fmt.Println("├─────────────────────┼──────────────────┼──────────────────┤")
	fmt.Printf("│ Exec/segundo        │ %16.2f │ %16.2f │\n", nativeStats.Throughput, yaegiStats.Throughput)
	fmt.Printf("│ Tempo Total         │ %16s │ %16s │\n",
		formatDuration(nativeStats.TotalTime),
		formatDuration(yaegiStats.TotalTime))
	fmt.Println("└─────────────────────┴──────────────────┴──────────────────┘")

	// Análise de overhead
	fmt.Println("\n======================================================================")
	fmt.Println("ANÁLISE DE OVERHEAD DO YAEGI")
	fmt.Println("======================================================================")

	overheadAvg := calculateOverhead(nativeStats.Avg, yaegiStats.Avg)
	overheadP95 := calculateOverhead(nativeStats.P95, yaegiStats.P95)
	overheadThroughput := ((nativeStats.Throughput - yaegiStats.Throughput) / yaegiStats.Throughput) * 100

	fmt.Printf("Overhead Médio:     %.2f%% mais lento\n", overheadAvg)
	fmt.Printf("Overhead P95:       %.2f%% mais lento\n", overheadP95)
	fmt.Printf("Throughput:         %.2f%% mais rápido no nativo\n\n", overheadThroughput)

	// Logging para o test output
	t.Logf("\n=== Performance Comparativa - %d execuções ===", numExecutions)
	t.Logf("Service Nativo - Min: %v | Avg: %v | P50: %v | P95: %v | Max: %v",
		nativeStats.Min, nativeStats.Avg, nativeStats.P50, nativeStats.P95, nativeStats.Max)
	t.Logf("Yaegi Script   - Min: %v | Avg: %v | P50: %v | P95: %v | Max: %v",
		yaegiStats.Min, yaegiStats.Avg, yaegiStats.P50, yaegiStats.P95, yaegiStats.Max)
	t.Logf("Overhead Médio: %.2f%% | Overhead P95: %.2f%%", overheadAvg, overheadP95)
}

// TestData representa um conjunto de dados de teste
type TestData struct {
	params    map[string]string
	operation functions.Operation
}

// generateTestData gera conjunto de dados de teste com valores variados
func generateTestData(count int) []TestData {
	testData := make([]TestData, count)

	for i := 0; i < count; i++ {
		// Valores aleatórios dentro dos intervalos especificados
		amount := 1000.0 + rand.Float64()*(100000.0-1000.0)          // 1000 a 100000
		interestRate := 0.01 + rand.Float64()*(0.05-0.01)            // 1% a 5%
		installments := 4 + rand.Intn(18-4+1)                        // 4 a 18

		testData[i] = TestData{
			params: map[string]string{
				"qtde_parcelas": fmt.Sprintf("%d", installments),
				"taxa_juros":    fmt.Sprintf("%.6f", interestRate),
			},
			operation: functions.Operation{
				Amount: amount,
			},
		}
	}

	return testData
}

// PerformanceStats armazena estatísticas de performance
type PerformanceStats struct {
	Min        time.Duration
	Avg        time.Duration
	P50        time.Duration
	P95        time.Duration
	Max        time.Duration
	StdDev     time.Duration
	Throughput float64
	TotalTime  time.Duration
}

// calculateStats calcula estatísticas de performance
func calculateStats(executionTimes []time.Duration) PerformanceStats {
	numExecutions := len(executionTimes)

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
	avgTime := totalTime / time.Duration(numExecutions)

	// Calcular P95 (percentil 95)
	p95Index := int(math.Ceil(float64(numExecutions)*0.95)) - 1
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

	// Calcular throughput
	totalTimeSeconds := totalTime.Seconds()
	throughput := float64(numExecutions) / totalTimeSeconds

	return PerformanceStats{
		Min:        minTime,
		Avg:        avgTime,
		P50:        p50Time,
		P95:        p95Time,
		Max:        maxTime,
		StdDev:     stdDev,
		Throughput: throughput,
		TotalTime:  totalTime,
	}
}

// printComparisonRow imprime uma linha da tabela comparativa
func printComparisonRow(metric string, nativeTime, yaegiTime time.Duration) {
	diff := float64(yaegiTime-nativeTime) / float64(nativeTime) * 100
	diffStr := fmt.Sprintf("%.2f%%", diff)
	if diff > 0 {
		diffStr = "+" + diffStr
	}

	fmt.Printf("│ %-19s │ %16s │ %16s │ %16s │\n",
		metric,
		formatDuration(nativeTime),
		formatDuration(yaegiTime),
		diffStr,
	)
}

// formatDuration formata duração para exibição
func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	} else if d < time.Millisecond {
		return fmt.Sprintf("%d µs", d.Microseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%.2f ms", float64(d.Microseconds())/1000.0)
	}
	return fmt.Sprintf("%.2f s", d.Seconds())
}

// calculateOverhead calcula o overhead percentual
func calculateOverhead(nativeTime, yaegiTime time.Duration) float64 {
	return (float64(yaegiTime-nativeTime) / float64(nativeTime)) * 100
}
