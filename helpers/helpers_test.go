package helpers

import (
	"fmt"

	"runtime"

	//"strings"
	"testing"
)

// Enable db debug mode to enable query logging
var debug = "false"

var Format = "\n---\nFAILURE\n---\ninput:    %q\nexpected: %q\noutput:   %q"

type Test struct {
	input    string
	expected string
}

func testLog() {
	pc, _, line, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()
	fmt.Printf("TEST LOG line %d - %s\n", line, name)
}

// TestPrices tests price conversion
func TestPrices(t *testing.T) {
	fmt.Println("\n---\nTESTING Prices\n---")

	var pence int
	var price string

	price = "£10.00"
	pence = PriceToCents(price)
	if pence != 1000 {
		t.Fatalf(Format, price, "1000", fmt.Sprintf("%d", pence))
	}
	price = CentsToPrice(int64(pence))
	if price != "10.00" {
		t.Fatalf(Format, price, "10.00", fmt.Sprintf("%d", pence))
	}

	price = "10"
	pence = PriceToCents(price)
	if pence != 1000 {
		t.Fatalf(Format, price, "1000", fmt.Sprintf("%d", pence))
	}
	price = CentsToPrice(int64(pence))
	if price != "10.00" {
		t.Fatalf(Format, price, "10.00", fmt.Sprintf("%d", pence))
	}

	price = "45"
	pence = PriceToCents(price)
	if pence != 4500 {
		t.Fatalf(Format, price, "4500", fmt.Sprintf("%d", pence))
	}
	price = CentsToPrice(int64(pence))
	if price != "45.00" {
		t.Fatalf(Format, price, "45.00", fmt.Sprintf("%d", pence))
	}

	price = "45.35"
	pence = PriceToCents(price)
	if pence != 4535 {
		t.Fatalf(Format, price, "4535", fmt.Sprintf("%d", pence))
	}

	price = CentsToPrice(int64(pence))
	if price != "45.35" {
		t.Fatalf(Format, price, "45.35", fmt.Sprintf("%d", pence))
	}

	price = "45.30"
	pence = PriceToCents(price)
	if pence != 4530 {
		t.Fatalf(Format, price, "4530", fmt.Sprintf("%d", pence))
	}

	price = CentsToPrice(int64(pence))
	if price != "45.30" {
		t.Fatalf(Format, price, "45.30", fmt.Sprintf("%d", pence))
	}

}
