package service

import (
	"fmt"
	"math"
	"poc-go-yaegi/functions"
	"strconv"
	"time"
)

// PriceCalculator é o service nativo que implementa cálculo da Tabela PRICE
type PriceCalculator struct{}

// NewPriceCalculator cria uma nova instância do calculador
func NewPriceCalculator() *PriceCalculator {
	return &PriceCalculator{}
}

// CalculateOperationStep calcula uma operação financeira usando a Tabela PRICE
// A Tabela PRICE é um sistema de amortização com parcelas fixas
func (c *PriceCalculator) CalculateOperationStep(params map[string]string, input functions.Operation) (*functions.Operation, error) {
	// Validar e extrair quantidade de parcelas
	qtdeParcelasStr, ok := params["qtde_parcelas"]
	if !ok {
		return nil, fmt.Errorf("parâmetro 'qtde_parcelas' não fornecido")
	}

	qtdeParcelas, err := strconv.Atoi(qtdeParcelasStr)
	if err != nil || qtdeParcelas <= 0 {
		return nil, fmt.Errorf("quantidade de parcelas inválida: %s", qtdeParcelasStr)
	}

	// Validar e extrair taxa de juros
	taxaJurosStr, ok := params["taxa_juros"]
	if !ok {
		return nil, fmt.Errorf("parâmetro 'taxa_juros' não fornecido")
	}

	taxaJuros, err := strconv.ParseFloat(taxaJurosStr, 64)
	if err != nil || taxaJuros < 0 {
		return nil, fmt.Errorf("taxa de juros inválida: %s", taxaJurosStr)
	}

	// Calcular coeficiente da Tabela PRICE
	// Fórmula: PMT = PV * [ i * (1+i)^n ] / [ (1+i)^n - 1 ]
	// Onde: PMT = valor da parcela, PV = valor presente, i = taxa, n = períodos
	var valorParcela float64
	if taxaJuros == 0 {
		// Se não há juros, divide igualmente
		valorParcela = input.Amount / float64(qtdeParcelas)
	} else {
		coeficiente := (math.Pow(1+taxaJuros, float64(qtdeParcelas)) * taxaJuros) /
			(math.Pow(1+taxaJuros, float64(qtdeParcelas)) - 1)
		valorParcela = input.Amount * coeficiente
	}

	// Preencher a estrutura da operação
	input.InstallmentsQty = qtdeParcelas
	input.Installments = make([]functions.Installment, qtdeParcelas)

	saldoDevedor := input.Amount
	totalPago := 0.0

	for i := 0; i < qtdeParcelas; i++ {
		// Calcular componentes da parcela
		juros := saldoDevedor * taxaJuros
		amortizacao := valorParcela - juros

		// Ajustar última parcela para evitar problemas de arredondamento
		if i == qtdeParcelas-1 {
			amortizacao = saldoDevedor
			valorParcela = amortizacao + juros
		}

		saldoDevedor -= amortizacao
		totalPago += valorParcela

		// Criar componentes da parcela
		components := []functions.Component{
			{Name: "Principal", Amount: amortizacao},
			{Name: "Juros", Amount: juros},
		}

		// Criar parcela
		input.Installments[i] = functions.Installment{
			Number:      i + 1,
			Amount:      valorParcela,
			TotalAmount: valorParcela,
			DueDate:     time.Now().AddDate(0, i+1, 0),
			Components:  components,
		}
	}

	input.TotalAmount = totalPago

	return &input, nil
}
