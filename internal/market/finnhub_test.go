package market

import (
	"context"
	"fmt"
	"testing"
)

func TestGetFX(t *testing.T) {
	client := NewFrankfurterFXProvider()
	rate, err := client.GetRate(context.Background(), "USD", "EUR")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(rate)

}
