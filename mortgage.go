package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
)

type PaymentPeriod float64

const (
	HELP_LOAN_AMOUNT            = "Loan Amount"
	HELP_PURCHASE               = "Purchase Price"
	HELP_AMORTIZATION_YEARS     = "Amortization Years"
	HELP_AMORTIZATION_MONTHS    = "Amortization Months"
	HELP_INTEREST_RATE          = "Annual Interest Rate"
	HELP_DOWNPAYMENT_PERCENTAGE = "Down Payment Percentage"
	HELP_PAYMENT_MONTHLY        = "Calculate Monthly Payment"
	HELP_PAYMENT_BIWEEKLY       = "Calculate Bi-Weekly Payment"
	HELP_PAYMENT_WEEKLY         = "Calculate Weekly Payment"
	HELP_PAYMENT_SCHEDULE       = "Print Payment Schedule"

	Monthly  PaymentPeriod = 1
	BiWeekly PaymentPeriod = 2
	Weekly   PaymentPeriod = 4
)

var (
	purchase           int
	loanAmount         float64
	amortizationYears  int
	amortizationMonths int
	interestRate       float64
	downPaymentPercent float64

	paymentMonthly  bool
	paymentBiweekly bool
	paymentWeekly   bool
	paymentSchedule bool
)

func init() {
	flag.IntVar(&purchase, "purchase", 0, HELP_PURCHASE)
	flag.IntVar(&purchase, "p", 0, HELP_PURCHASE)
	flag.Float64Var(&loanAmount, "loan", 0, HELP_LOAN_AMOUNT)
	flag.Float64Var(&loanAmount, "l", 0, HELP_LOAN_AMOUNT)

	flag.IntVar(&amortizationYears, "amortization", 0, HELP_AMORTIZATION_YEARS)
	flag.IntVar(&amortizationYears, "ay", 0, HELP_AMORTIZATION_YEARS)
	flag.IntVar(&amortizationMonths, "amortization-months", 0, HELP_AMORTIZATION_MONTHS)
	flag.IntVar(&amortizationMonths, "am", 0, HELP_AMORTIZATION_MONTHS)

	flag.Float64Var(&interestRate, "interest-rate", 3, HELP_INTEREST_RATE)
	flag.Float64Var(&interestRate, "i", 3, HELP_INTEREST_RATE)

	flag.Float64Var(&downPaymentPercent, "down-percent", 0, HELP_DOWNPAYMENT_PERCENTAGE)
	flag.Float64Var(&downPaymentPercent, "d", 0, HELP_DOWNPAYMENT_PERCENTAGE)

	flag.BoolVar(&paymentMonthly, "monthly", true, HELP_PAYMENT_MONTHLY)
	flag.BoolVar(&paymentBiweekly, "biweekly", false, HELP_PAYMENT_BIWEEKLY)
	flag.BoolVar(&paymentWeekly, "weekly", false, HELP_PAYMENT_WEEKLY)
	flag.BoolVar(&paymentSchedule, "schedule", false, HELP_PAYMENT_SCHEDULE)
}

func main() {
	flag.Parse()

	if purchase > 0 && downPaymentPercent > 0 {
		loanAmount = float64(purchase) * (1 - downPaymentPercent/100)
	}

	if amortizationMonths > 0 {
		amortizationYears = amortizationMonths * 12
	}

	mortgage := NewMortgage(loanAmount, interestRate, amortizationYears)
	fmt.Println(mortgage)

	/*
		switch {
		case paymentMonthly:
			payment := mortgage.PaymentMonthly()
			fmt.Printf("Payment (Montly): %0.2f\n", payment)
		case paymentBiweekly:
			payment := mortgage.PaymentBiWeekly()
			fmt.Printf("Payment (BiWeekly): %0.2f\n", payment)
		case paymentWeekly:
			payment := mortgage.PaymentWeekly()
			fmt.Printf("Payment (Weekly): %0.2f\n", payment)
		}
	*/

	fmt.Printf("Payment (Montly):   %0.2f\n", mortgage.Payment(Monthly))
	fmt.Printf("Payment (BiWeekly): %0.2f\n", mortgage.Payment(BiWeekly))
	fmt.Printf("Payment (Weekly):   %0.2f\n", mortgage.Payment(Weekly))

	out := os.Stdout

	var payments float64
	var interest float64
	var brokeEven bool

	fmt.Fprintln(out, "Year | Month |    Value    |  Payment  | Principal | Interest")
	for _, payment := range mortgage.Schedule(Monthly) {

		if !brokeEven && payment.Principal >= payment.Interest {
			fmt.Fprintln(out, "------------------------- Break Even -------------------------")
		}

		fmt.Fprintf(out, "% 4d | % 5d | % 11.2f | % 9.2f | % 9.2f | % 8.2f \n", payment.Year, payment.Month, payment.Value, payment.Payment, payment.Principal, payment.Interest)

		if !brokeEven && payment.Principal >= payment.Interest {
			fmt.Fprintln(out, "------------------------- Break Even -------------------------")
			brokeEven = true
		}

		interest += payment.Interest
		payments += payment.Principal
	}
	fmt.Fprintf(out, "Payments:% 12.2f\n", payments)
	fmt.Fprintf(out, "Interest:% 12.2f\n", interest)
	fmt.Fprintf(out, "Total:   % 12.2f\n", payments+interest)
}

type Mortgage struct {
	Principal          float64
	InterestRateAnnual float64
	AmortizationYears  int
}

func NewMortgage(loanAmount, interestRate float64, amortizationYears int) *Mortgage {
	return &Mortgage{
		Principal:          loanAmount,
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

func payment(P, r float64, N int) float64 {
	return (r * P) / (1 - math.Pow(1+r, -1*float64(N)))
}

func (t Mortgage) LoanValueMonths(months int) float64 {
	r := t.InterestRateAnnual / 12 / 100
	P := t.Principal
	N := t.AmortizationYears * 12
	c := t.Payment(Monthly)
	return loanValue(c, P, r, N)
}

//(1+r)^N P - (1+r)^N-1 * (c/r)
func loanValue(c, P, r float64, n int) float64 {
	exp := math.Pow(1+r, float64(n))
	return exp*P - (exp-1)*c/r
}

type Payment struct {
	Value     float64
	Payment   float64
	Principal float64
	Interest  float64
	Year      int
	Month     int
}

func (t Payment) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Year / Month: %02d/%02d\n", t.Year, t.Month)
	fmt.Fprintf(buf, "Value:        %0.2f\n", t.Value)
	fmt.Fprintf(buf, "Payment:      %0.2f\n", t.Payment)
	fmt.Fprintf(buf, "Principal:    %0.2f\n", t.Principal)
	fmt.Fprintf(buf, "Interest:     %0.2f\n", t.Interest)
	return buf.String()
}

func (t Mortgage) Schedule(pp PaymentPeriod) []Payment {
	r := t.InterestRateAnnual / 12 / 100
	P := t.Principal
	N := t.AmortizationYears * 12
	c := t.Payment(pp)

	payments := make([]Payment, N+1)

	principal := P
	for i := 0; i < N+1; i++ {
		value := loanValue(c, P, r, i)

		payments[i] = Payment{
			Year:      (i / 12) + 1,
			Month:     (i % 12) + 1,
			Value:     value,
			Payment:   c,
			Principal: principal - value,
			Interest:  c - (principal - value),
		}
		principal = value
	}
	return payments
}

func (t Mortgage) String() string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Principal:              %0.2f\n", t.Principal)
	fmt.Fprintf(buf, "Interest Rate (Annual): %0.2f\n", t.InterestRateAnnual)
	fmt.Fprintf(buf, "Amortization (Years):   %d\n", t.AmortizationYears)
	return buf.String()
}
