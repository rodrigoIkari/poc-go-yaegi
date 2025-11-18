package functions

import "time"

// Operation representa uma operação financeira completa
type Operation struct {
	Amount          float64       // Valor principal financiado
	TotalAmount     float64       // Valor total a ser pago (principal + juros)
	InstallmentsQty int           // Quantidade de parcelas
	Installments    []Installment // Lista de parcelas da operação
}

// Installment representa uma parcela individual da operação
type Installment struct {
	Number      int         // Número da parcela (1, 2, 3...)
	Amount      float64     // Valor total da parcela
	TotalAmount float64     // Valor total acumulado até esta parcela
	DueDate     time.Time   // Data de vencimento da parcela
	Components  []Component // Componentes da parcela (principal, juros, etc)
}

// Component representa um componente de uma parcela (ex: principal, juros, IOF)
type Component struct {
	Name   string  // Nome do componente
	Amount float64 // Valor do componente
}
