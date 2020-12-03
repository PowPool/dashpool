package dashcoin

import (
	"fmt"
	"testing"
)

func TestGetCoinBaseScriptHex1(t *testing.T) {
	scriptHex, _ := GetCoinBaseScriptHex("XiB2rj7PdESyaxJVsnmjhXf9D9bYJjX7ob")
	fmt.Println("coinbaser script:", scriptHex)
}

func TestGetCoinBaseScriptHex2(t *testing.T) {
	scriptHex, _ := GetCoinBaseScriptHex("034a452d21d26c60076a30bf6701666b30d57ac09c2ff07f34e52cdba13796645d")
	fmt.Println("coinbaser script:", scriptHex)
}
