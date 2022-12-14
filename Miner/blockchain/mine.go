package blockchain

import (
	"BlockchainInGo/transaction"
	"BlockchainInGo/utils"
	"bytes"
	"encoding/hex"
	"log"
)

// RunMine 挖矿
func (bc *BlockChain) RunMine(address []byte) Block {
	tp := CreateTransactionPool()
	ok, tx := bc.VerifyTransactions(tp.PubTx)
	for !ok {
		println(len(tp.PubTx), " transactions")
		log.Printf("%x falls in verification", tx.ID)
		tp.DeleteInvalidTransactions(tx)
		tp = CreateTransactionPool()
		ok, tx = bc.VerifyTransactions(tp.PubTx)
	}
	basetx := transaction.BaseTx(address)
	tp.PubTx = append(tp.PubTx, basetx)
	candidateBlock := CreateBlock(bc.LastHash, tp.PubTx)
	//println("---------------------------------------1---------------------------------------")
	//fmt.Printf("hash:%x\n", candidateBlock.Hash)
	//fmt.Println(candidateBlock.PrevHash)
	//println("---------------------------------------2---------------------------------------")
	RemoveTransactionPoolFile()
	SetDifficulty(candidateBlock)
	return *candidateBlock
}

func SetDifficultyForTests(block *Block) {
	lastBlock := ReadLB()
	block.Difficulty = lastBlock.Difficulty
}

func SetDifficulty(block *Block) {
	lastBlock := ReadLB()
	past_d := ReadLB().Difficulty
	//times := utils.ReadBlockTime("time.txt")
	// 每十个块调整一次难度
	if block.Index%10 == 0 {
		times := ReadTime()
		interval := utils.AverageInterval(times)
		if interval < 2 {
			block.Difficulty = lastBlock.Difficulty + 1
		} else if interval > 4 {
			block.Difficulty = lastBlock.Difficulty - 1
		} else {
			block.Difficulty = lastBlock.Difficulty
		}
		println("difficult changed from ", past_d, " to ", block.Difficulty)
	} else {
		block.Difficulty = lastBlock.Difficulty
	}

}

// VerifyTransactions 对于交易池里的每一个交易，都应该先验证每一个交易输入是否合法：
//				1.是否是重复的交易输入
//				2.该输入是否包含在对应地址的UTXO中
//				3.签名是否正确
//最后还要 		4.判断整个交易的输入钱数是否等于输出钱数
func (bc *BlockChain) VerifyTransactions(txs []*transaction.Transaction) (bool, *transaction.Transaction) {
	if len(txs) == 0 {
		return true, nil
	}

	spentOutputs := make(map[string]int)
	for _, tx := range txs {
		//感觉这两行是错的，应该变成循环里面的那两行
		//pubKey := tx.Inputs[0].PubKey
		//unspentOutputs := bc.FindUnspentTransactions(pubKey)
		inputAmount := 0
		OutputAmount := 0

		for _, input := range tx.Inputs {
			pubKey := input.PubKey
			unspentOutputs := bc.FindUnspentTransactions(pubKey)
			outidx, ok := spentOutputs[hex.EncodeToString(input.TxID)] //1.是否是重复的交易输入
			if ok && int64(outidx) == input.OutIdx {
				return false, tx
			}
			ok, amount := isInputRight(unspentOutputs, input) //2.该输入是否包含在对应地址的UTXO中
			if !ok {
				return false, tx
			}
			inputAmount += amount
			spentOutputs[hex.EncodeToString(input.TxID)] = int(input.OutIdx) //记录已经花掉的钱
		}
		for _, output := range tx.TxOutput {
			OutputAmount += output.Value
		}
		if inputAmount != OutputAmount { //4.判断整个交易的输入钱数是否等于输出钱数
			return false, tx
		}
		if !tx.Verify() { //3.签名是否正确
			return false, tx
		}
	}
	return true, nil
}

func isInputRight(txs []transaction.Transaction, in transaction.TxInput) (bool, int) {
	for _, tx := range txs { //txs是某地址所有的UTXO，循环的逻辑是如果这些UTXO中包含当前验证交易的交易输入（也就是它想花的钱属于它能花的钱）则返回true和这笔钱的值
		if bytes.Equal(tx.ID, in.TxID) {
			return true, tx.TxOutput[in.OutIdx].Value
		}
	}
	return false, 0
}
