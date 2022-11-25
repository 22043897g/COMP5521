package test

import (
	"BlockchainInGo/blockchain"
	"BlockchainInGo/constcoe"
	"BlockchainInGo/utils"
	"BlockchainInGo/wallet"
	"bytes"
	"fmt"
	"os"
)

// SendRefName 根据钱包别名打钱
func SendRefName(from, to string, amount int) {
	refList := wallet.LoadRefList()
	fromAddress, err := refList.FindRef(from)
	utils.Handle(err)
	toAddress, err := refList.FindRef(to)
	utils.Handle(err)
	Send(fromAddress, toAddress, amount)
}

func Send(from, to string, amount int) {
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	fromWallet := wallet.LoadWallet(from)
	tx, ok := chain.CreateTransaction(fromWallet.PublicKey, utils.Address2PubHash([]byte(to)), amount, fromWallet.PrivateKey)
	if !ok {
		fmt.Println("Failed to create transaction")
		return
	}
	tp := blockchain.CreateTransactionPool() //注意这个函数是创建空池并加载硬盘存的池
	tp.AddTransaction(tx)
	tp.SaveFile()
	fmt.Println("Success!")
}

func CreateGenBlock(address string) *blockchain.Block {
	Gen := blockchain.GenesisBlock(utils.Address2PubHash([]byte(address)))
	fmt.Println("Finished creating GenBlock")
	return Gen
}

// CreateGenBlockRefName 根据钱包别名创建创世区块
func CreateGenBlockRefName(refname string) *blockchain.Block {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	block := CreateGenBlock(address)
	return block
}

// CreateBlockChainRefName 根据钱包别名创建区块链
func CreateBlockChainRefName(refname string, block *blockchain.Block) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	CreateBlockChain(address, block)
}

func CreateBlockChain(address string, block *blockchain.Block) {
	newChain := blockchain.InitBlockChain(utils.Address2PubHash([]byte(address)), block)
	newChain.Database.Close()
	fmt.Println("Finished creating blockchain,and the owner is :", address)
}

//func Mine(address []byte) blockchain.Block {
//	chain := blockchain.ContinueBlockChain()
//	defer chain.Database.Close()
//	block := chain.RunMine(address)
//	writeErr := ioutil.WriteFile("the weirdest fucking bug.txt", block.PrevHash, 0666)
//	if writeErr != nil {
//		panic("sth wrong")
//	}
//	//println("---------------------------------------3---------------------------------------")
//	//fmt.Printf("hash:%x\n", block.Hash)
//	//fmt.Println(block.PrevHash)
//	//println("---------------------------------------4---------------------------------------")
//	fmt.Println("Finish Mining")
//	return block
//}

func BlockChainInfo() {
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	iterator := chain.Iterator()
	ogprevhash := chain.BackOgPrevHash()
	for {
		block := iterator.Next()
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Printf("Difficulty:%d\n", block.Difficulty)
		fmt.Printf("Height:%d\n", block.Index)
		fmt.Printf("Timestamp:%d\n", block.Timestamp)
		fmt.Printf("Previous hash:%x\n", block.PrevHash)
		fmt.Printf("Number of transactions:%d\n", len(block.Transactions))
		fmt.Printf("hash:%x\n", block.Hash)
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Println()
		if bytes.Equal(block.PrevHash, ogprevhash) {
			break
		}
	}
}

func Balance(address string) {
	chain := blockchain.ContinueBlockChain() //读出区块链
	defer chain.Database.Close()
	wlt := wallet.LoadWallet(address)
	balance, _ := chain.FindUTXOs(wlt.PublicKey)
	fmt.Printf("Address:%s,balance:%d\n", address, balance)
}

func BalanceRefName(refname string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	println("===================Name:", refname, "=========================")
	Balance(address)
}

func WalletAddress(refname string) string {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	return address
}

func WalletsList() {
	refList := wallet.LoadRefList()
	for address, _ := range *refList {
		wlt := wallet.LoadWallet(address)
		fmt.Println("------------------------------------------------------")
		fmt.Printf("Wallet address:%x\n\n", address)
		fmt.Printf("Public Key:%x\n", wlt.PrivateKey)
		fmt.Printf("Reference Name:%s\n", (*refList)[address])
		fmt.Println("------------------------------------------------------")
		fmt.Println()
	}
}

func Init() *blockchain.Block {
	os.RemoveAll("tmp/blocks")
	os.RemoveAll("D:/tmp/wallets")
	os.RemoveAll("D:/tmp/ref_list")
	os.Mkdir("tmp/blocks", os.ModePerm)
	os.Mkdir("D:/tmp/wallets", os.ModePerm)
	os.Mkdir("D:/tmp/ref_list", os.ModePerm)
	wallet.CreateWallet("A")
	wallet.CreateWallet("B")
	wallet.CreateWallet("C")
	wallet.CreateWallet("D")
	block := CreateGenBlockRefName(constcoe.Refname)
	return block
}

func ShowWallets() {
	reflist := *wallet.LoadRefList()
	for _, v := range reflist {
		BalanceRefName(v)
	}
}
