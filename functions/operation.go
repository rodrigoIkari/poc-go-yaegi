package functions

import "time"

type Operation struct {
	Amount          float64
	TotalAmount     float64
	InstallmentsQty int
	Installments    []Installment
}

type Installment struct {
	Number      int
	Amount      float64
	TotalAmount float64
	DueDate     time.Time
	Components  []Component
}

type Component struct {
	Name   string
	Amount float64
}
