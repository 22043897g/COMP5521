package cli

import (
	"BlockchainInGo/blockchain"
	"BlockchainInGo/utils"
	"BlockchainInGo/wallet"
	"bytes"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct {
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Welcome to Leo Cao's tiny blockchain system, usage is as follows:")
	fmt.Println("---------------------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println("All you need is to first create a wallet.")
	fmt.Println("And then you can use the wallet address to create a blockchain and declare the owner.")
	fmt.Println("Make transactions to expand the blockchain.")
	fmt.Println("In addition, don't forget to run mine function after transatcions are collected.")
	fmt.Println("---------------------------------------------------------------------------------------------------------------------------------------------------------")
	fmt.Println("createwallet -refname REFNAME                       ----> Creates and save a wallet. The refname is optional.")
	fmt.Println("walletinfo -refname NAME -address Address           ----> Print the information of a wallet. At least one of the refname and address is required.")
	fmt.Println("walletsupdate                                       ----> Registrate and update all the wallets (especially when you have added an existed .wlt file).")
	fmt.Println("walletslist                                         ----> List all the wallets found (make sure you have run walletsupdate first).")
	fmt.Println("createblockchain -refname NAME -address ADDRESS     ----> Creates a blockchain with the owner you input (address or refname).")
	fmt.Println("balance -refname NAME -address ADDRESS              ----> Back the balance of a wallet using the address (or refname) you input.")
	fmt.Println("blockchaininfo                                      ----> Prints the blocks in the chain.")
	fmt.Println("send -from FROADDRESS -to TOADDRESS -amount AMOUNT  ----> Make a transaction and put it into candidate block.")
	fmt.Println("sendbyrefname -from NAME1 -to NAME2 -amount AMOUNT  ----> Make a transaction and put it into candidate block using refname.")
	fmt.Println("mine                                                ----> Mine and add a block to the chain.")
	fmt.Println("---------------------------------------------------------------------------------------------------------------------------------------------------------")
}

// createBlockChain 创建区块链
func (cli *CommandLine) createBlockChain(address string) {
	newChain := blockchain.InitBlockChain(utils.Address2PubHash([]byte(address)))
	newChain.Database.Close()
	fmt.Println("Finished creating blockchain,and the owner is :", address)
}

// 查询余额
func (cli *CommandLine) balance(address string) {
	chain := blockchain.ContinueBlockChain() //读出区块链
	defer chain.Database.Close()
	wlt := wallet.LoadWallet(address)
	balance, _ := chain.FindUTXOs(wlt.PublicKey)
	fmt.Printf("Address:%s,balance:%d\n", address, balance)
}

// 创建新钱包（需要存入钱包列表）
func (cli *CommandLine) createWallet(refname string) {
	newWallet := wallet.NewWallet()
	newWallet.Save()
	reflist := wallet.LoadRefList()
	// 在钱包列表中加入新钱包并存储
	reflist.BindRef(string(newWallet.Address()), refname)
	reflist.Save()
	fmt.Println("Succeed in creating wallet.")
}

// 根据钱包地址读取钱包并输出其信息
func (cli *CommandLine) walletInfo(address string) {
	wlt := wallet.LoadWallet(address)
	refList := wallet.LoadRefList()
	fmt.Printf("Wallet address:%x\n\n", wlt.Address())
	fmt.Printf("Public Key:%x\n", wlt.PrivateKey)
	fmt.Printf("Reference Name:%s\n", (*refList)[address])
}

// 读取钱包列表并根据钱包别名找到钱包地址(每个钱包单独存储并以钱包地址命名，通过别名地址的映射可以找到地址并读取钱包）
func (cli *CommandLine) walletInfoRefName(refname string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	cli.walletInfo(address)
}

// 本地创建的钱包会被直接添加到钱包列表，所以Update仅在有其他结点拷贝来的钱包时使用
func (cli *CommandLine) walletUpdate() {
	refList := wallet.LoadRefList()
	refList.Update()
	refList.Save()
	fmt.Println("Succeed in updating wallets.")
}

// 遍历钱包列表并输出钱包信息（输出信息那块应该也可以直接调用walletInfo）
func (cli *CommandLine) walletsList() {
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

// 根据钱包别名打钱
func (cli *CommandLine) sendRefName(from, to string, amount int) {
	refList := wallet.LoadRefList()
	fromAddress, err := refList.FindRef(from)
	utils.Handle(err)
	toAddress, err := refList.FindRef(to)
	utils.Handle(err)
	cli.send(fromAddress, toAddress, amount)
}

// 根据钱包别名创建区块链
func (cli *CommandLine) createBlockChainRefName(refname string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	cli.createBlockChain(address)
}

//
func (cli *CommandLine) balanceRefName(refname string) {
	refList := wallet.LoadRefList()
	address, err := refList.FindRef(refname)
	utils.Handle(err)
	cli.balance(address)
}

// getBlockChainInfo 遍历区块并显示信息
func (cli *CommandLine) getBlockChainInfo() {
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	iterator := chain.Iterator()
	ogprevhash := chain.BackOgPrevHash()
	for {
		block := iterator.Next()
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Printf("Difficulty:%d\n", block.Difficulty)
		fmt.Printf("Target:%d\n", utils.BytesToInt64(block.Target))
		fmt.Printf("Height:%d\n", block.Index)
		fmt.Printf("Timestamp:%d\n", block.Timestamp)
		fmt.Printf("Previous hash:%x\n", block.PrevHash)
		fmt.Printf("Transactions:%v\n", block.Transactions)
		fmt.Printf("hash:%x\n", block.Hash)
		fmt.Printf("Pow: %s\n", strconv.FormatBool(block.ValidPow()))
		fmt.Println("--------------------------------------------------------------------------------------------------------------")
		fmt.Println()
		if bytes.Equal(block.PrevHash, ogprevhash) {
			break
		}
	}
}

// send 打钱
func (cli *CommandLine) send(from, to string, amount int) {
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

// mine 开挖
func (cli *CommandLine) mine(address []byte) {
	chain := blockchain.ContinueBlockChain()
	defer chain.Database.Close()
	fmt.Println("Start Mining")
	chain.RunMine(address)
	fmt.Println("Finish Mining")
}

func (cli *CommandLine) validdateArgs() {
	if len(os.Args) < 2 { //Args是个数组，[0]存储程序名称（cli.go）,[1]存储命令名称，[2]到[n-1]存储输入的参数
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	walletInfoCmd := flag.NewFlagSet("walletinfo", flag.ExitOnError)
	walletsUpdateCmd := flag.NewFlagSet("walletsupdate ", flag.ExitOnError)
	walletsListCmd := flag.NewFlagSet("walletslist", flag.ExitOnError)
	createBlockChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	getBlockChainInfoCmd := flag.NewFlagSet("blockchaininfo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendByRefNameCmd := flag.NewFlagSet("sendbyrefname", flag.ExitOnError)
	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)

	createWalletRefName := createWalletCmd.String("refname", "", "The refname of the wallet, and this is optimal")
	walletInfoRefName := walletInfoCmd.String("refname", "", "The refname of the wallet")
	walletInfoAddress := walletInfoCmd.String("address", "", "The address of the wallet")
	createBlockChainOwner := createBlockChainCmd.String("address", "", "The address refer to the owner of blockchain")
	createBlockChainByRefNameOwner := createBlockChainCmd.String("refname", "", "The name refer to the owner of blockchain")
	balanceAddress := balanceCmd.String("address", "", "Who needs to get balance amount")
	balanceRefName := balanceCmd.String("refname", "", "Who needs to get balance amount")
	sendByRefNameFrom := sendByRefNameCmd.String("from", "", "Source refname")
	sendByRefNameTo := sendByRefNameCmd.String("to", "", "Destination refname")
	sendByRefNameAmount := sendByRefNameCmd.Int("amount", 0, "Amount to send")
	sendFromAddress := sendCmd.String("from", "", "Source address")
	sendToAddress := sendCmd.String("to", "", "Destination address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "walletinfo":
		err := walletInfoCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "walletsupdate":
		err := walletsUpdateCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "walletslist":
		err := walletsListCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "createblockchain":
		err := createBlockChainCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "balance":
		err := balanceCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "blockchaininfo":
		err := getBlockChainInfoCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "sendbyrefname":
		err := sendByRefNameCmd.Parse(os.Args[2:])
		utils.Handle(err)

	case "mine":
		err := mineCmd.Parse(os.Args[2:])
		utils.Handle(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(*createWalletRefName)
	}

	if walletInfoCmd.Parsed() {
		if *walletInfoAddress == "" {
			if *walletInfoRefName == "" {
				walletInfoCmd.Usage()
				runtime.Goexit()
			} else {
				cli.walletInfoRefName(*walletInfoRefName)
			}
		} else {
			cli.walletInfo(*walletInfoAddress)
		}
	}

	if walletsUpdateCmd.Parsed() {
		cli.walletUpdate()
	}

	if walletsListCmd.Parsed() {
		cli.walletsList()
	}

	if createBlockChainCmd.Parsed() {
		if *createBlockChainOwner == "" {
			if *createBlockChainByRefNameOwner == "" {
				createBlockChainCmd.Usage()
				runtime.Goexit()
			} else {
				cli.createBlockChainRefName(*createBlockChainByRefNameOwner)
			}
		} else {
			cli.createBlockChain(*createBlockChainOwner)
		}
	}

	if balanceCmd.Parsed() {
		if *balanceAddress == "" {
			if *balanceRefName == "" {
				balanceCmd.Usage()
				runtime.Goexit()
			} else {
				cli.balanceRefName(*balanceRefName)
			}
		} else {
			cli.balance(*balanceAddress)
		}
	}

	if sendByRefNameCmd.Parsed() {
		if *sendByRefNameFrom == "" || *sendByRefNameTo == "" || *sendByRefNameAmount <= 0 {
			sendByRefNameCmd.Usage()
			runtime.Goexit()
		}
		cli.sendRefName(*sendByRefNameFrom, *sendByRefNameTo, *sendByRefNameAmount)
	}

	if sendCmd.Parsed() {
		if *sendFromAddress == "" || *sendToAddress == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*sendFromAddress, *sendToAddress, *sendAmount)
	}

	if getBlockChainInfoCmd.Parsed() {
		cli.getBlockChainInfo()
	}

	if mineCmd.Parsed() {
		//cli.mine()
	}
}
