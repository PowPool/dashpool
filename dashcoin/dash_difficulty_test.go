package dashcoin

import (
	"fmt"
	"testing"
)

func TestNBits2Target(t *testing.T) {
	targetGenesis := NBits2Target(GENESISNBITS)
	fmt.Println("base 10, targetGenesis:", targetGenesis.Text(10))
	fmt.Printf("base 16, targetGenesis: %064s", targetGenesis.Text(16))
}
