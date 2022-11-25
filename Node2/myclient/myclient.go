package myclient

import (
	"BlockchainInGo/blockchain"
	"BlockchainInGo/constcoe"
	"BlockchainInGo/proto"
	"BlockchainInGo/test"
	"BlockchainInGo/transaction"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

var addrs = []string{"0.0.0.0:8001"} //, "0.0.0.0:8002", "0.0.0.0:8003", "0.0.0.0:8004"}

func NewTransaction(client proto.BlockchainServiceClient, tx transaction.Transaction) {
	res, err := client.NewTransaction(context.Background(), &proto.TransactionsRequest{
		BelongHash: nil,
		ID:         tx.ID,
	})
	if err != nil {
		log.Fatalf("Could not broadcast wallet: %v\n", err)
	}
	println(res.TR)
}

func NewOutputs(client proto.BlockchainServiceClient, tx transaction.Transaction) {
	//log.Println("BroadCastOutputs was invoked")
	stream, err := client.NewOutPuts(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet %v\n", err)
	}
	outputs := []*proto.OutputsRequest{}

	for i, output := range tx.TxOutput {
		pOutput := proto.OutputsRequest{
			BelongId:   tx.ID,
			Index:      int64(i),
			Value:      int64(output.Value),
			HashPubKey: output.HashPubKey,
		}
		outputs = append(outputs, &pOutput)
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

func NewInputs(client proto.BlockchainServiceClient, tx transaction.Transaction) {
	//log.Println("BroadCastInputs was invoked")
	stream, err := client.NewInPuts(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet %v\n", err)
	}
	inputs := []*proto.InputsRequest{}
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

func GetBlock(client proto.BlockchainServiceClient, height int64) {
	res, err := client.GetBlock(context.Background(), &proto.GetBlockRequest{
		Height:  height,
		Address: constcoe.Address,
	})
	if err != nil {
		log.Fatalf("Could not broadcast wallet: %v\n", err)
	}
	println(res.WR)
}

func BroadcastCreator(client proto.BlockchainServiceClient, name string) {
	_, err := client.CreateBy(context.Background(), &proto.CreateByRequest{
		Name: name,
	})
	if err != nil {
		log.Fatalf("Could not broadcast wallet: %v\n", err)
	}
}

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

func BroadcastNewEnd(client proto.BlockchainServiceClient) {
	//log.Println("BroadCastEnd was invoked")
	r, err := client.NewEnd(context.Background(), &proto.EndRequest{
		EndFlag: true,
	})
	if err != nil {
		log.Fatalf("Could not broadcast block: %v\n", err)
	}
	log.Println(r.ER)
}

func BroadcastGenBlock(block blockchain.Block, i int) {
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

func GetBlocks() {
	conn, err := grpc.Dial(addrs[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)

	var CHeight int64
	if !blockchain.BlockchainExist() {
		CHeight = 0
	} else {
		CHeight = blockchain.ReadLB().Index
	}
	println("CHeight:", CHeight)

	GetBlock(client, CHeight)
}

func BroadCastNewTransaction(form, to string, amount int) {
	conn, err := grpc.Dial(addrs[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)
	test.SendRefName(form, to, amount)
	txp := blockchain.CreateTransactionPool()
	println("len:", len(txp.PubTx))
	tx := *txp.PubTx[0]
	NewTransaction(client, tx)
	NewInputs(client, tx)
	NewOutputs(client, tx)
	BroadcastNewEnd(client)
	blockchain.RemoveTransactionPoolFile()
}

func DoubleSpending(from, to1, to2 string, amount1, amount2 int) {
	BroadCastNewTransaction(from, to1, amount1)
	BroadCastNewTransaction(from, to2, amount2)
}

func StartClient() {
	GetBlocks()
	//DoubleSpending("C", "B", "A", 20, 20)
	//BroadCastNewTransaction("A", "C", 60)
	//BroadCastNewTransaction("A", "D", 70)
	//BroadCastNewTransaction("C", "A", 80)
	//BroadCastNewTransaction("A", "D", 15)
	//fmt.Printf("hash in block:%x\n", block.Hash)
	//fmt.Printf("Client_hash:%x\n", block.CalculateHash())

}
