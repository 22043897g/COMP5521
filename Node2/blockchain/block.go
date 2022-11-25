package blockchain

import (
	"BlockchainInGo/merkletree"
	"BlockchainInGo/transaction"
	"BlockchainInGo/utils"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/go-redis/redis"
	"io/ioutil"
	"strconv"
	"time"
)

var rdb *redis.Client

// redis初始化
func initRedis() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379", // 指定
		Password: "",
		DB:       0, // redis一共16个库，指定其中一个库即可
	})
	_, err = rdb.Ping().Result()
	return
}

func RedisInit() {
	err := initRedis()
	if err != nil {
		fmt.Printf("connect redis failed! err : %v\n", err)
		return
	}
	//fmt.Println("redis连接成功！")
}

type Block struct {
	Index        int64 //高度
	Timestamp    int64
	Hash         []byte
	PrevHash     []byte
	Difficulty   int64
	Target       []byte //方便其他结点快速验证nonce是否正确
	Nonce        int64
	Transactions []*transaction.Transaction //记录交易信息
	MTree        *merkletree.MerkleTree
}

// SetHash 计算并设置区块hash
func (b *Block) SetHash() {
	// information 将区块的各属性连接起来，用以求哈希值，最后一个参数是连接时的分隔符，此处取空
	data := bytes.Join([][]byte{utils.Int64ToBytes(b.Index), utils.Int64ToBytes(b.Timestamp), b.PrevHash, b.Target, utils.Int64ToBytes(b.Nonce), b.BackTransactionSummary(), b.MTree.RootNode.Data}, []byte{})
	// sum256返回的是sha256的校验和，[32]byte形式，无法直接用b.Hash接收
	hash := sha256.Sum256(data)
	b.Hash = hash[:]
}

// CalculateHash 计算哈希，用于client发送时的测试
func (b *Block) CalculateHash() []byte {
	// information 将区块的各属性连接起来，用以求哈希值，最后一个参数是连接时的分隔符，此处取空
	data := bytes.Join([][]byte{utils.Int64ToBytes(b.Index), utils.Int64ToBytes(b.Timestamp), b.PrevHash, b.Target, utils.Int64ToBytes(b.Nonce), b.BackTransactionSummary(), b.MTree.RootNode.Data}, []byte{})
	// sum256返回的是sha256的校验和，[32]byte形式，无法直接用b.Hash接收
	hash := sha256.Sum256(data)
	return hash[:]
}

func WriteLB(block Block) {
	inputfile := "LB.txt"
	outputfile := inputfile
	bs, readError := ioutil.ReadFile(inputfile)
	if readError != nil {
		panic("sth wrong")
	}
	bs = block.Serialize()
	// perm是读写权限
	writeErr := ioutil.WriteFile(outputfile, bs, 0666)
	if writeErr != nil {
		panic("sth wrong")
	}
}

func ReadLB() Block {
	RedisInit()
	block := Block{}
	res, _ := rdb.Get("difficulty").Result()
	difficulty, _ := strconv.ParseInt(res, 10, 64)
	res2, _ := rdb.Get("index").Result()
	index, _ := strconv.ParseInt(res2, 10, 64)
	block.Index = index
	block.Difficulty = difficulty
	return block
}

// CreateBlock 创建区块
func CreateBlock(prevhash []byte, txs []*transaction.Transaction) *Block {
	var block Block
	lastBlock := ReadLB()
	if bytes.Equal(prevhash, []byte("And there was light.")) {
		block = Block{1, time.Now().Unix(), []byte{}, prevhash, 8, []byte{}, 0, txs, merkletree.CreateMerkleTree(txs)}
		WriteLB(block)
	} else {
		height := lastBlock.Index
		block = Block{height + 1, time.Now().Unix(), []byte{}, prevhash, 0, []byte{}, 0, txs, merkletree.CreateMerkleTree(txs)}
	}
	block.Target = block.GetTarget()
	block.Nonce = block.FindNonce()
	block.SetHash()
	return &block
}

// GenesisBlock 创建第一笔交易，并将其放入创世区块
func GenesisBlock(address []byte) *Block {
	tx := transaction.BaseTx(address)
	genesis := CreateBlock([]byte("And there was light."), []*transaction.Transaction{tx})
	genesis.SetHash()
	return genesis
}

// BackTransactionSummary 将所有交易信息整合为一个byte slice,以帮助SetHash函数和GetBase4Nonce函数进行序列化
func (b *Block) BackTransactionSummary() []byte {
	txIDs := make([][]byte, 0)
	for _, tx := range b.Transactions {
		txIDs = append(txIDs, tx.ID)
	}
	summary := bytes.Join(txIDs, []byte{})
	return summary
}

// ValidPow 计算hash并与target比较以验证nonce是否合法(注意需要计算hash不能直接使用区块中的hash）
func (b *Block) ValidPow() bool {
	targer := utils.BytesToInt64(b.Target)
	data := b.GetBase4Nonce(b.Nonce)
	hash := sha256.Sum256(data)
	hash_b := hash[:]
	hash_64 := utils.BytesToInt64(hash_b)
	if hash_64 <= targer {
		return true
	} else {
		return false
	}
}

// Badger只能序列化存储,所以要有序列化和反序列化函数

// Serialize 将数据序列化
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	utils.Handle(err)
	return res.Bytes()
}

// DeSerialize 反序列化
func DeSerialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	utils.Handle(err)
	return &block
}
