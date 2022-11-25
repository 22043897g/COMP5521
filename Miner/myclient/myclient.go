package myclient

import (
	"BlockchainInGo/addresses"
	"BlockchainInGo/blockchain"
	"BlockchainInGo/constcoe"
	"BlockchainInGo/proto"
	"BlockchainInGo/test"
	"BlockchainInGo/utils"
	"BlockchainInGo/wallet"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func BroadcastCreator(client proto.BlockchainServiceClient, name string) {
	_, err := client.CreateBy(context.Background(), &proto.CreateByRequest{
		Name: name,
	})
	if err != nil {
		log.Fatalf("Could not broadcast wallet: %v\n", err)
	}
}

// BroadCastBlock 广播区块
func BroadCastBlock(client proto.BlockchainServiceClient, block blockchain.Block) {
	//log.Println("BroadCastBlock was invoked")
	_, err := client.Block(context.Background(), &proto.BlockRequest{
		Index:      block.Index,
		Timestamp:  block.Timestamp,
		Hash:       block.Hash,
		PrevHash:   block.PrevHash,
		Difficulty: block.Difficulty,
		Target:     block.Target,
		Nonce:      block.Nonce,
	})
	if err != nil {
		log.Fatalf("Could not broadcast block: %v\n", err)
	}
	//log.Println("Block broadcast:", r.BR)
}

// BroadCastTransactions 广播交易
func BroadCastTransactions(client proto.BlockchainServiceClient, block blockchain.Block) {
	//log.Println("BroadCastTransactions was invoked")
	stream, err := client.Transactions(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet %v\n", err)
	}
	txs := []*proto.TransactionsRequest{}
	for i := 0; i < len(block.Transactions); i++ {
		ptx := proto.TransactionsRequest{
			BelongHash: block.Hash,
			ID:         block.Transactions[i].ID,
		}
		txs = append(txs, &ptx)
	}
	for _, tx := range txs {
		//log.Printf("Sending transaction:%x\n", tx.ID)
		stream.Send(tx)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet :%v\n", err)
	}
	//log.Println("Transactions result:", res)
}

// BroadCastInputs 广播交易输入
func BroadCastInputs(client proto.BlockchainServiceClient, block blockchain.Block) {
	//log.Println("BroadCastInputs was invoked")
	stream, err := client.Inputs(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet %v\n", err)
	}
	inputs := []*proto.InputsRequest{}
	for _, tx := range block.Transactions {
		for i, input := range tx.Inputs {
			pInput := proto.InputsRequest{
				BelongId: tx.ID,
				Index:    int64(i),
				TxID:     input.TxID,
				OutIdx:   input.OutIdx,
				PubKey:   input.PubKey,
				Sig:      input.Sig,
			}
			inputs = append(inputs, &pInput)
		}
	}
	for _, inp := range inputs {
		//log.Printf("Sending input:%x\n,from:%x", inp.Index, inp.BelongId)
		stream.Send(inp)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet :%v\n", err)
	}
	//log.Println("Transactions result:", res)
}

// BroadCastOutputs 广播交易输出
func BroadCastOutputs(client proto.BlockchainServiceClient, block blockchain.Block) {
	//log.Println("BroadCastOutputs was invoked")
	stream, err := client.Outputs(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet %v\n", err)
	}
	outputs := []*proto.OutputsRequest{}
	for _, tx := range block.Transactions {
		for i, output := range tx.TxOutput {
			pOutput := proto.OutputsRequest{
				BelongId:   tx.ID,
				Index:      int64(i),
				Value:      int64(output.Value),
				HashPubKey: output.HashPubKey,
			}
			outputs = append(outputs, &pOutput)
		}
	}
	for _, outp := range outputs {
		//log.Printf("Sending output:%x\n,from:%x", outp.Index, outp.BelongId)
		stream.Send(outp)
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet :%v\n", err)
	}
	//log.Println("Transactions result:", res)
}

// BroadcastEnd 广播传输完毕信号，对方收到后可以根据收到的区块、交易、交易输入输出重新组合出区块，验证后加入区块链
func BroadcastEnd(client proto.BlockchainServiceClient) {
	//log.Println("BroadCastEnd was invoked")
	r, err := client.End(context.Background(), &proto.EndRequest{
		EndFlag: true,
	})
	if err != nil {
		log.Fatalf("Could not broadcast block: %v\n", err)
	}
	log.Println(r.ER)
}

// BroadcastGenBlock 广播创世区块（创世区块比较特殊，要单独广播，对方收到后根据创世区块生成区块链）
func BroadcastGenBlock(block blockchain.Block, i int) {
	addrs := addresses.ReadAllAddress()
	conn, err := grpc.Dial(addrs[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)
	BroadcastCreator(client, constcoe.Refname)
	BroadCastBlock(client, block)
	BroadCastTransactions(client, block)
	BroadCastInputs(client, block)
	BroadCastOutputs(client, block)
	BroadcastEnd(client)
}

func LongMiningTest() {
	address := test.WalletAddress(constcoe.Refname)
	for i := 0; i < 30; i++ {
		test.Tx1()
		txp := blockchain.CreateTransactionPool()
		if len(txp.PubTx) == 0 {
			//println("Mining...")
			time.Sleep(3 * time.Second)
			continue
		}
		block := test.Mine(utils.Address2PubHash([]byte(address)))
		// 这prehash传过来必出问题，print都报错，在test.Mine里print一点事都没有。而且hash无论在哪里print也都一点事没有，看了俩小时也不知道为什么只能改成用读写文件传递了
		prehash, readError := ioutil.ReadFile("the weirdest fucking bug.txt")
		if readError != nil {
			panic("sth wrong")
		}
		block.PrevHash = prehash
		addrs := addresses.ReadAllAddress()
		for i := 0; i < len(addrs); i++ {
			conn, err := grpc.Dial(addrs[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to connect: %v\n", err)
			}
			defer conn.Close()
			// 创建新客户端
			client := proto.NewBlockchainServiceClient(conn)
			BroadCastBlock(client, block)
			BroadCastTransactions(client, block)
			BroadCastInputs(client, block)
			BroadCastOutputs(client, block)
			BroadcastEnd(client)
			//fmt.Printf("hash in block:%x\n", block.Hash)
			//fmt.Printf("Client_hash:%x\n", block.CalculateHash())
			//test.ShowWallets()
		}
	}
}

// Init 清空所有钱包和区块链信息，生成miner钱包，生成创世区块，广播给所有结点
func Init(addr string) {
	os.RemoveAll("tmp/blocks")
	os.RemoveAll("D:/tmp/wallets")
	os.RemoveAll("D:/tmp/ref_list")
	os.Mkdir("tmp/blocks", os.ModePerm)
	os.Mkdir("D:/tmp/wallets", os.ModePerm)
	os.Mkdir("D:/tmp/ref_list", os.ModePerm)
	wallet.CreateWallet(addr)
}

func BlockchainInit() {
	block := test.CreateGenBlockRefName(constcoe.Refname)
	// 广播创世区块
	addrs := addresses.ReadAllAddress()
	for i := 0; i < len(addrs); i++ {
		BroadcastGenBlock(*block, i)
	}
	//test.ShowWallets()
}

func BeginMining() {
	address := test.WalletAddress(constcoe.Refname)
	for {
		txp := blockchain.CreateTransactionPool()
		if len(txp.PubTx) == 0 {
			//println("Mining...")
			time.Sleep(3 * time.Second)
			continue
		}
		block := test.Mine(utils.Address2PubHash([]byte(address)))
		// 这prehash传过来必出问题，print都报错，在test.Mine里print一点事都没有。而且hash无论在哪里print也都一点事没有，看了俩小时也不知道为什么只能改成用读写文件传递了
		prehash, readError := ioutil.ReadFile("the weirdest fucking bug.txt")
		if readError != nil {
			panic("sth wrong")
		}
		block.PrevHash = prehash
		addrs := addresses.ReadAllAddress()
		for i := 0; i < len(addrs); i++ {
			conn, err := grpc.Dial(addrs[i], grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to connect: %v\n", err)
			}
			defer conn.Close()
			// 创建新客户端
			client := proto.NewBlockchainServiceClient(conn)
			BroadCastBlock(client, block)
			BroadCastTransactions(client, block)
			BroadCastInputs(client, block)
			BroadCastOutputs(client, block)
			BroadcastEnd(client)
			//fmt.Printf("hash in block:%x\n", block.Hash)
			//fmt.Printf("Client_hash:%x\n", block.CalculateHash())
			//test.ShowWallets()
		}
	}
}

func InitAll() {
	addresses.PortInit()
	Init("A")
	BlockchainInit()
}

func StartClient() {
	//InitAll()
	//LongMiningTest()
	BeginMining()

}
