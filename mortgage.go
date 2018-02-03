package mortgage

import (
	"bytes"
	"fmt"
	"math"
)

const (
	Monthly  PaymentPeriod = 1
	BiWeekly PaymentPeriod = 2
	Weekly   PaymentPeriod = 4
)

type PaymentPeriod float64

type Mortgage struct {
	Principal          float64
	InterestRateAnnual float64
	AmortizationYears  int
}

func NewMortgage(principal, interestRate float64, amortizationYears int) *Mortgage {
	return &Mortgage{
		Principal:          principal,
		InterestRateAnnual: interestRate,
		AmortizationYears:  amortizationYears,
	}
}

func (t Mortgage) Payment(pp PaymentPeriod) float64 {
	r := t.InterestRateAnnual / 12 / 100
	P := t.Principal
	N := t.AmortizationYears * 12

	switch pp {
	case BiWeekly:
		return payment(P, r, N) / float64(BiWeekly)
	case Weekly:
		return payment(P, r, N) / float64(Weekly)
	default:
		return payment(P, r, N) / float64(Monthly)
	}
}

func (t Mortgage) NumPayments(pp PaymentPeriod) int {
	switch pp {
	case BiWeekly:
		return t.AmortizationYears * 26
	case Weekly:
		return t.AmortizationYears * 52
	default:
		return t.AmortizationYears * 12
	}
}

func payment(P, r float64, N int) float64 {
	return (r * P) / (1 - math.Pow(1+r, -1*float64(N)))
}

func (t Mortgage) LoanValue(pp PaymentPeriod, periods int) float64 {
	r := t.InterestRateAnnual / 12 / 100
	P := t.Principal
	c := t.Payment(pp)
	return loanValue(c, P, r, periods)
}

//(1+r)^N P - (1+r)^N-1 * (c/r)
func loanValue(c, P, r float64, n int) float64 {
	exp := math.Pow(1+r, float64(n))
	return exp*P - (exp-1)*c/r
}

type Payment struct {
	Period       int
	Year         int
	Number       int
	Value        float64
	Payment      float64
	ExtraPayment float64
	Principal    float64
	Interest     float64
	Tax          float64
}

func (t Payment) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Period:    %d\n", t.Period)
	fmt.Fprintf(buf, "Year:      %d\n", t.Year)
	fmt.Fprintf(buf, "Number:    %d\n", t.Number)
	fmt.Fprintf(buf, "Value:     %0.2f\n", t.Value)
	fmt.Fprintf(buf, "Payment:   %0.2f\n", t.Payment)
	fmt.Fprintf(buf, "Principal: %0.2f\n", t.Principal)
	fmt.Fprintf(buf, "Interest:  %0.2f\n", t.Interest)
	return buf.String()
}

func (t Mortgage) Schedule(pp PaymentPeriod, extraPayment float64) []Payment {
	r := t.InterestRateAnnual / (12 * float64(pp)) / 100
	principal := t.Principal
	N := t.NumPayments(pp)

	payment := t.Payment(pp)

	payments := make([]Payment, 0, N+1)

	for period := 1; period < N+1 && principal > 0; period++ {

		interest := r * principal
		principal -= payment + extraPayment - interest

		if principal < 0 {
			payment = payment + principal
			principal = 0
		}

		var number int
		var year int

		switch pp {
		case Weekly:
			year = (period / 52) + 1
			number = (period % 52) + 1
		case BiWeekly:
			year = (period / 26) + 1
			number = (period % 26) + 1
		default:
			year = ((period - 1) / 12) + 1
			number = ((period - 1) % 12) + 1
		}

		payments = append(payments, Payment{
			Number:       number,
			Year:         year,
			Period:       period,
			Value:        principal,
			Payment:      payment,
			ExtraPayment: extraPayment,
			Principal:    payment - interest,
			Interest:     interest,
		})
	}

	return payments
}

func (t Mortgage) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Principal:              % 11.2f\n", t.Principal)
	fmt.Fprintf(buf, "Interest Rate (Annual): % 11.2f\n", t.InterestRateAnnual)
	fmt.Fprintf(buf, "Amortization (Years):   % 11d\n", t.AmortizationYears)
	fmt.Fprintf(buf, "Montly Payment (%d):   % 11.2f\n", t.NumPayments(Monthly), t.Payment(Monthly))
	fmt.Fprintf(buf, "BiWeekly Payment (%d): % 11.2f\n", t.NumPayments(BiWeekly), t.Payment(BiWeekly))
	fmt.Fprintf(buf, "Weekly Payment (%d):  % 11.2f\n", t.NumPayments(Weekly), t.Payment(Weekly))
	return buf.String()
}
