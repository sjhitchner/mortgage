package main

import (
	"flag"
	"fmt"
	"github.com/sjhitchner/mortgage"
	"io"
	"os"
)

const (
	HELP_PRINCIPAL              = "Loan Principal"
	HELP_PURCHASE               = "Purchase Price"
	HELP_AMORTIZATION_YEARS     = "Amortization Years"
	HELP_AMORTIZATION_MONTHS    = "Amortization Months"
	HELP_INTEREST_RATE          = "Annual Interest Rate"
	HELP_DOWNPAYMENT_PERCENTAGE = "Down Payment Percentage"
	HELP_EXTRA_PAYMENT          = "Extra Payment Per Period"
	HELP_TAX_RATE               = "Tax Rate Percentage"
	HELP_PAYMENT_MONTHLY        = "Calculate Monthly Payment"
	HELP_PAYMENT_BIWEEKLY       = "Calculate Bi-Weekly Payment"
	HELP_PAYMENT_WEEKLY         = "Calculate Weekly Payment"
	HELP_PAYMENT_SCHEDULE       = "Print Payment Schedule"
	HELP_OUTPUT_FILE            = "Output to File"
)

var (
	purchase           float64
	principal          float64
	amortizationYears  int
	amortizationMonths int
	interestRate       float64
	downPaymentPercent float64
	taxRate            float64
	extraPayment       float64

	outputFile string

	paymentMonthly  bool
	paymentBiweekly bool
	paymentWeekly   bool
	paymentSchedule bool

	number int
)

func init() {
	flag.StringVar(&outputFile, "output", "", HELP_OUTPUT_FILE)
	flag.StringVar(&outputFile, "f", "", HELP_OUTPUT_FILE)

	flag.Float64Var(&principal, "principal", 0, HELP_PRINCIPAL)
	flag.Float64Var(&principal, "p", 0, HELP_PRINCIPAL)

	flag.IntVar(&amortizationYears, "amortization", 0, HELP_AMORTIZATION_YEARS)
	flag.IntVar(&amortizationYears, "ay", 0, HELP_AMORTIZATION_YEARS)
	flag.IntVar(&amortizationMonths, "amortization-months", 0, HELP_AMORTIZATION_MONTHS)
	flag.IntVar(&amortizationMonths, "am", 0, HELP_AMORTIZATION_MONTHS)

	flag.Float64Var(&interestRate, "interest-rate", 3, HELP_INTEREST_RATE)
	flag.Float64Var(&interestRate, "i", 3, HELP_INTEREST_RATE)

	flag.Float64Var(&purchase, "purchase", 0, HELP_PURCHASE)
	flag.Float64Var(&purchase, "s", 0, HELP_PURCHASE)
	flag.Float64Var(&downPaymentPercent, "down-percent", 0, HELP_DOWNPAYMENT_PERCENTAGE)
	flag.Float64Var(&downPaymentPercent, "d", 0, HELP_DOWNPAYMENT_PERCENTAGE)

	flag.Float64Var(&extraPayment, "extra-payment", 0, HELP_EXTRA_PAYMENT)
	flag.Float64Var(&extraPayment, "ep", 0, HELP_EXTRA_PAYMENT)

	flag.Float64Var(&taxRate, "tax-rate", 0, HELP_TAX_RATE)
	flag.Float64Var(&taxRate, "t", 0, HELP_TAX_RATE)

	flag.BoolVar(&paymentMonthly, "monthly", true, HELP_PAYMENT_MONTHLY)
	flag.BoolVar(&paymentBiweekly, "biweekly", false, HELP_PAYMENT_BIWEEKLY)
	flag.BoolVar(&paymentWeekly, "weekly", false, HELP_PAYMENT_WEEKLY)
	flag.BoolVar(&paymentSchedule, "schedule", false, HELP_PAYMENT_SCHEDULE)

	flag.IntVar(&number, "n", 0, HELP_AMORTIZATION_MONTHS)
}

func main() {
	flag.Parse()

	if purchase > 0 && downPaymentPercent > 0 {
		principal = purchase * (1 - downPaymentPercent/100)
	}

	if amortizationMonths > 0 {
		amortizationYears = amortizationMonths * 12
	}

	m := mortgage.NewMortgage(principal, interestRate, amortizationYears)
	fmt.Println(m)

	out, err := Output()
	CheckError(err)
	defer out.Close()

	paymentPeriod := PaymentPeriod()

	var payments float64
	var interest float64
	var brokeEven bool

	if paymentSchedule {

		fmt.Fprintln(out, "Period | Year | Num |  Payment  |  Extra  | Principal | Interest |    Value    ")
		fmt.Fprintf(out, "       |      |     |           |         |           |          | % 11.2f\n", principal)

		for i, payment := range m.Schedule(paymentPeriod, extraPayment) {

			if !brokeEven && payment.Principal >= payment.Interest {
				fmt.Fprintln(out, "------------------------- Break Even -------------------------")
			}

			fmt.Fprintf(out, "% 6d | % 3d  | % 3d | % 9.2f | % 7.2f | % 9.2f | % 8.2f | % 11.2f\n",
				payment.Period,
				payment.Year,
				payment.Number,
				payment.Payment,
				payment.ExtraPayment,
				payment.Principal,
				payment.Interest,
				payment.Value,
			)

			if !brokeEven && payment.Principal >= payment.Interest {
				fmt.Fprintln(out, "------------------------- Break Even -------------------------")
				brokeEven = true
			}

			interest += payment.Interest
			payments += payment.Principal + payment.ExtraPayment

			if i > number && number != 0 {
				break
			}
		}

		fmt.Fprintf(out, "Payments:% 12.2f\n", payments)
		fmt.Fprintf(out, "Interest:% 12.2f\n", interest)
		fmt.Fprintf(out, "Total:   % 12.2f\n", payments+interest)
	}
}

func PaymentPeriod() mortgage.PaymentPeriod {
	switch {
	case paymentBiweekly:
		return mortgage.BiWeekly
	case paymentWeekly:
		return mortgage.Weekly
	default:
		return mortgage.Monthly
	}
}

func Output() (io.WriteCloser, error) {
	if outputFile == "" {
		return os.Stdout, nil
	}
	return os.Create(outputFile)
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(-1)
	}
}
