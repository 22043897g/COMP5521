package main

import (
	"BlockchainInGo/blockchain"
	"BlockchainInGo/constcoe"
	"BlockchainInGo/merkletree"
	"BlockchainInGo/myclient"
	"BlockchainInGo/proto"
	"BlockchainInGo/test"
	"BlockchainInGo/transaction"
	"BlockchainInGo/wallet"
	"bytes"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"net"
	"strconv"
)

type blockrequest struct {
	height  int64
	address string
}

var creator string
var block blockchain.Block
var requester blockrequest

type Server struct {
	proto.BlockchainServiceServer
}

func Basic(b *blockchain.Block, in *proto.BlockRequest) {
	b.Difficulty = in.Difficulty
	b.Timestamp = in.Timestamp
	b.Nonce = in.Nonce
	b.Target = in.Target
	b.Index = in.Index
	b.Hash = in.Hash
	b.PrevHash = in.PrevHash
}

func (*Server) GetBlock(ctx context.Context, in *proto.GetBlockRequest) (*proto.CreateByResponse, error) {
	// 将所有需要发送的区块放入blocks
	bc := blockchain.ContinueBlockChain()
	iter := bc.Iterator()
	blocks := []blockchain.Block{}
	block := *iter.Next()
	for {
		if block.Index != in.Height {
			blocks = append(blocks, block)
			block = *iter.Next()
		} else {
			break
		}
	}
	//发送
	conn, err := grpc.Dial(in.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer conn.Close()
	client := proto.NewBlockchainServiceClient(conn)
	for i := len(blocks); i > 0; i-- {
		myclient.BroadCastBlock(client, blocks[i])
		myclient.BroadCastTransactions(client, blocks[i])
		myclient.BroadCastInputs(client, blocks[i])
		myclient.BroadCastOutputs(client, blocks[i])
		myclient.BroadcastEnd(client)
	}
	return &proto.CreateByResponse{WR: "Blocks sent"}, nil
}

func (*Server) CreateBy(ctx context.Context, in *proto.CreateByRequest) (*proto.CreateByResponse, error) {
	log.Printf("Createby was invoked with %s\n", in.Name)
	creator = in.Name
	return &proto.CreateByResponse{WR: "Received name: " + in.Name}, nil
}

func (*Server) Block(ctx context.Context, in *proto.BlockRequest) (*proto.BlockResponse, error) {
	log.Printf("Transactions was invoked with %d\n", in.Index)
	Basic(&block, in)
	return &proto.BlockResponse{BR: "Received block" + strconv.FormatInt(in.Index, 10)}, nil
}

func (*Server) Transactions(stream proto.BlockchainService_TransactionsServer) error {
	// 清空transaction
	txs := []*transaction.Transaction{}
	block.Transactions = txs
	log.Println("Transactions was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Transaction transmission finished," + strconv.FormatInt(count, 10) + " transactions have been received."
			return stream.SendAndClose(&proto.TxsResponse{
				TR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易:%x,它属于%x\n", req.ID, req.BelongHash)
		tx := transaction.Transaction{
			ID:       req.ID,
			Inputs:   nil,
			TxOutput: nil,
		}
		block.Transactions = append(block.Transactions, &tx)
	}
}

func (*Server) Inputs(stream proto.BlockchainService_InputsServer) error {
	log.Println("Inputs was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Inputs transmission finished," + strconv.FormatInt(count, 10) + " inputs have been received."
			return stream.SendAndClose(&proto.InputsResponse{
				IR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易输入:%x,它属于%x\n", req.Index, req.BelongId)
		input := transaction.TxInput{
			TxID:   req.TxID,
			OutIdx: req.OutIdx,
			PubKey: req.PubKey,
			Sig:    req.Sig,
		}
		for i := 0; i < len(block.Transactions); i++ {
			tx := block.Transactions[i]
			if bytes.Equal(tx.ID, req.BelongId) {
				tx.Inputs = append(tx.Inputs, input)
			}
		}
	}
}
func (*Server) Outputs(stream proto.BlockchainService_OutputsServer) error {
	log.Println("Outputs was invoked")
	count := int64(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			res := "Inputs transmission finished," + strconv.FormatInt(count, 10) + " outputs have been received."
			return stream.SendAndClose(&proto.OutputsResponse{
				OR: res,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream :%v\n", err)
		}
		// 注意这个输出在服务端
		count += 1
		//fmt.Printf("收到交易输出:%x,它属于%x\n", req.Index, req.BelongId)
		output := transaction.TxOutput{
			Value:      int(req.Value),
			HashPubKey: req.HashPubKey,
		}
		for i := 0; i < len(block.Transactions); i++ {
			tx := block.Transactions[i]
			if bytes.Equal(tx.ID, req.BelongId) {
				tx.TxOutput = append(tx.TxOutput, output)
			}
		}
	}
}

func (*Server) End(ctx context.Context, in *proto.EndRequest) (*proto.EndResponse, error) {
	log.Printf("End was invoked")
	judge := ""
	if block.Index == 1 {
		test.CreateBlockChainRefName(creator, &block)
		judge = "passed."
	} else {
		root := merkletree.CreateMerkleTree(block.Transactions)
		block.MTree = root
		flag := VerifyBlock(block)
		if flag {
			judge = "passed."
			bc := blockchain.ContinueBlockChain()
			bc.AddBlock(&block)
			blockchain.WriteLB(block)
			//utils.WriteBlockTime(strconv.FormatInt(time.Now().Unix(), 10))
			//utils.WriteBlockTime("\n")
		} else {
			judge = "failed."
		}

		//fmt.Printf("Server_hash:%x", tools.VerifyHash(block))
		//fmt.Printf("Txs_hash:%x\n", tools.VerifyTxsHash(block))
	}
	test.ShowWallets()
	return &proto.EndResponse{
		ER: "Node2:Block " + strconv.FormatInt(block.Index, 10) + " received,verification " + judge,
	}, nil
}

func VerifyBlock(block blockchain.Block) bool {
	if !block.ValidPow() {
		return false
	} else {
		hash := block.CalculateHash()
		if !bytes.Equal(hash, block.Hash) {
			return false
		}
	}
	return true
}

func main() {
	wallet.CreateWallet("C")
	lis, err := net.Listen("tcp", constcoe.Address)
	if err != nil {
		log.Fatalf("Failed to listen on: %v\n", err)
	}
	log.Printf("Listening on %s\n", constcoe.Address)
	//创建服务器
	server := grpc.NewServer()
	//注册(后面那个参数包含所有*Server接收器)
	proto.RegisterBlockchainServiceServer(server, &Server{})
	if err = server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve:%v\n", err)
	}
}
