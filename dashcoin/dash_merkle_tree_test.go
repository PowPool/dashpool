package dashcoin

import (
	"fmt"
	"testing"
)

func TestGetMerkleBranchHexFromTxIdsWithoutCoinBase(t *testing.T) {
	//"aec5799ca150e2e25efd3e78aa649701f8a02444f50a75ba4938602f127f4700"
	//"7fc898f6166ce4bedc49bd6514d64940b9c28bc0647ed5b62e4a22c58da4ab1a"
	//"c73d7fd5cea0ec81e26b215344cde90759c715b867591a0406cb882d0584603d"
	//"fbd50735e27899aa17f59d38c1e5fa1cde431848e7a76490c5524b065a082b3a"
	//"95575f8562b810290d0f637fbee1fa52e8d829aae8862680d339fa99b40331b5"
	txIdsHexWithCoinBase := []string{
		"aec5799ca150e2e25efd3e78aa649701f8a02444f50a75ba4938602f127f4700",
		"18ac6e72a78a062fd24d74d7e86e3605ab96df036c24ce792d2a471a38631b26",
		"a23426065899dc8b8995f8e2baaabb5c423eaa432034c8d6c54c54dfc2e91739",
		"bcc145c463a1ed5198a9ef84644b173ce0db298a1331f23f7befb42115da9f49",
		"b57db2bea71e2cbdcb4044ec5ff346f5497b91a76311506078c86ac0de452158",
		"3b1e3d37c2c472e2696cdad3e2ce0794152a66d59d83a6acf1b16f4499d71f5f",
		"2ad70917e86b8018d3c17604c2916eb88e18185c315a74d88d7fb8fe0fee5d63",
		"17b1d3806ff59c286c12d0621cd417e0e47620567af9503d606b4dbb88491170",
		"3c82e59bd388f08e05e1e2dc4314a561383157eb6dfa8676b5686b5e916fa970",
		"0ab031ea9298b421a3a670fc9b0c7b7ca6b76601e2f9e347393174cd737859b0",
		"9b79043a62ebce175aef786c009d96043cfea9a2f77e0c1243f76b3f6c841df2",
		"a5dcda71dd39ec1185d80fa5573a44c67a88e95d1529884e78a9207693ff1300",
		"34c337b6d5b602ba892026fc42ecb8d6256cb1b9839fe9fd0cbbe50eab143723",
		"eacab53b71b916f1b49a7154f5afa6f0f6c9a3b58770d288aebf442e9bb0ae35",
		"87c36b4657253492c33f483c42fd8e6f65758f2fddce1ee13cc6468ace6aa064",
		"6b8b0adc9198a355c77b43028b3b36cb266de984fd4da077a3ab8aa17489528b",
		"50e66a0de514fc0365afd845cd785ec9b57fec01171f61eaf92fad159242579d",
		"6fd61260acdc0c08713914ae4f3b7a959216b3f2759d1591c0872d065bb4c09f",
		"3238992ec1cd9daef3f066e8d5d88370ab7fe5fe31f782dc7dbf8df7e2a3c3ce",
		"d0e9717782e05a89c67d7c97a3293d3e93377de8bffb7fb578e0d4d671e567f7"}
	merkleBranch, _ := GetMerkleBranchHexFromTxIdsWithoutCoinBase(txIdsHexWithCoinBase)
	fmt.Println("merkleBranch len:", len(merkleBranch))
	fmt.Println("merkleBranch:", merkleBranch)
}

func TestGetMerkleRootHexFromTxIdsWithCoinBase(t *testing.T) {
	//"6960d019913a8958642415b92836304a2f39275df60bfbc30e65020489ac2b64"
	txIdsHexWithCoinBase := []string{
		"a9c02cb69f753ef724110f7a0b95724492ded6ac1333f22424de0b8eafdb35a2", //coinbase tx id
		"aec5799ca150e2e25efd3e78aa649701f8a02444f50a75ba4938602f127f4700",
		"18ac6e72a78a062fd24d74d7e86e3605ab96df036c24ce792d2a471a38631b26",
		"a23426065899dc8b8995f8e2baaabb5c423eaa432034c8d6c54c54dfc2e91739",
		"bcc145c463a1ed5198a9ef84644b173ce0db298a1331f23f7befb42115da9f49",
		"b57db2bea71e2cbdcb4044ec5ff346f5497b91a76311506078c86ac0de452158",
		"3b1e3d37c2c472e2696cdad3e2ce0794152a66d59d83a6acf1b16f4499d71f5f",
		"2ad70917e86b8018d3c17604c2916eb88e18185c315a74d88d7fb8fe0fee5d63",
		"17b1d3806ff59c286c12d0621cd417e0e47620567af9503d606b4dbb88491170",
		"3c82e59bd388f08e05e1e2dc4314a561383157eb6dfa8676b5686b5e916fa970",
		"0ab031ea9298b421a3a670fc9b0c7b7ca6b76601e2f9e347393174cd737859b0",
		"9b79043a62ebce175aef786c009d96043cfea9a2f77e0c1243f76b3f6c841df2",
		"a5dcda71dd39ec1185d80fa5573a44c67a88e95d1529884e78a9207693ff1300",
		"34c337b6d5b602ba892026fc42ecb8d6256cb1b9839fe9fd0cbbe50eab143723",
		"eacab53b71b916f1b49a7154f5afa6f0f6c9a3b58770d288aebf442e9bb0ae35",
		"87c36b4657253492c33f483c42fd8e6f65758f2fddce1ee13cc6468ace6aa064",
		"6b8b0adc9198a355c77b43028b3b36cb266de984fd4da077a3ab8aa17489528b",
		"50e66a0de514fc0365afd845cd785ec9b57fec01171f61eaf92fad159242579d",
		"6fd61260acdc0c08713914ae4f3b7a959216b3f2759d1591c0872d065bb4c09f",
		"3238992ec1cd9daef3f066e8d5d88370ab7fe5fe31f782dc7dbf8df7e2a3c3ce",
		"d0e9717782e05a89c67d7c97a3293d3e93377de8bffb7fb578e0d4d671e567f7"}
	merkleRoot, _ := GetMerkleRootHexFromTxIdsWithCoinBase(txIdsHexWithCoinBase)
	fmt.Println("merkleRoot:", merkleRoot)
}
