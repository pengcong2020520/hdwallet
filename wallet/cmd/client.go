package cmd

import (
	"fmt"
	"log"
	"os"
	"flag"
	"math/big"
	"context"
	"io/ioutil"
	"strings"
	"encoding/json"

	"go_code/hdwallet/hd"
	"go_code/hdwallet/hdKeystore"
	"go_code/hdwallet/erc20"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/tyler-smith/go-bip39"
	"github.com/howeyc/gopass"


)


type CLI struct {
	DataPath string
	NetWorkUrl string
	TokenFile string 
}

type TokenConfig struct {
	Symbol string `json: "symbol"`
	Addr string `json: "addr"` // ....?
}

func NewCLI(path, url, tokenfile string) *CLI {
	return &CLI{
		DataPath : path,
		NetWorkUrl : url,
		TokenFile : tokenfile,
	}
}

//提供帮助 
func (c *CLI) Help() {
	fmt.Println("create new wallet : hdwallet.exe createwallet -name + ACCOUNT_NAME")
	fmt.Println("get wallet balance : hdwallet.exe getbalance -name + ACCOUNT_NAME --for get wallet balance")
	fmt.Println("transfer to : hdwallet.exe transfer -name + ACCOUNT_NAME -to + Addr -value + VALUE --for transfer value to toaddr")
	fmt.Println("add token to : hdwallet.exe addtoken -address Contract_Addr --addtoken to address")
	fmt.Println("send token to : hdwallet.exe sendtoken -name + ACCOUNT_NAME -symbol + SYMBOL -to + ADDRESS -amount + AMOUNT --sendtoken amount to address")
	fmt.Println("get token balance : hdwallet.exe tokenbalance -name + ACCOUNT -symbol + SYMBOL")
}

//参数检测
func (c *CLI) Valid() {
	if len(os.Args) < 2 {
		c.Help()
		os.Exit(-1)
	}
}

func (c *CLI) Run() {
	//运行前先进行检测  check
	c.Valid()
	//解析命令行
	//1. 分类设定
		//（1） 创建账户
	walletFlag := flag.NewFlagSet("createwallet", flag.ExitOnError)  //创建一个Flag
		//（2） 查询余额
	balanceFlag := flag.NewFlagSet("getbalance", flag.ExitOnError)
		//（3） 账户转账
	transferFlag := flag.NewFlagSet("transfer", flag.ExitOnError)
		// (4)  支持token
	addtokenFlag := flag.NewFlagSet("addtoken", flag.ExitOnError)
		// (5)  转账token
	sendtokenFlag := flag.NewFlagSet("sendtoken", flag.ExitOnError)
		// (6)  查询token余额
	tokenbalanceFlag := flag.NewFlagSet("tokenbalance", flag.ExitOnError)

	//2. 设定要解析的具体参数
			//String用指定的名称、默认值、使用信息注册一个string类型flag。
			//返回一个保存了该flag的值的指针。
		// (1) createwallet
	walletAccount := walletFlag.String("name", "123", "ACCOUNT_NAME") 
		// (2) getbalance
	walletBalanceAccount := balanceFlag.String("name", "", "ACCOUNT_NAME")
		// (3) transfer
	transferName := transferFlag.String("name", "", "ACCOUNT_NAME")
	transferToAddr := transferFlag.String("to", "", "TOADDR")
	transferValue := transferFlag.Int64("value", 0, "VALUE")
		// (4) add token
	//addtokenAccount := addtokenFlag.String("name", "", "ACCOUNT_NAME")
	addtokenAddr := addtokenFlag.String("address", "", "Contract_Addr")
		// (5) send token 
	sendtokenName := sendtokenFlag.String("name", "", "ACCOUNT_NAME")
	sendtokenSymbol := sendtokenFlag.String("symbol", "", "SYMBOL")
	sendtokenToAddr := sendtokenFlag.String("to", "", "ADDRESS")
	sendtokenAmount := sendtokenFlag.Int64("amount", 0, "AMOUNT")
		// (6) get token balance
	tokenbalanceName := tokenbalanceFlag.String("name", "", "ACCOUNT_NAME")
	tokenbalanceSymbol := tokenbalanceFlag.String("symbol", "", "SYMBOL")


		//接收命令行输入的参数
	switch os.Args[1] {
		case "createwallet" :
			//解析createwallet这个参数
			err := walletFlag.Parse(os.Args[2:])
			if err != nil {
				fmt.Println("Create wallet Parse error")
				return 
			}
		case "getbalance" :
			err := balanceFlag.Parse(os.Args[2:])
			if err != nil {
				fmt.Println("get balance Parse error")
				return 
			}
		case "transfer" :
			err := transferFlag.Parse(os.Args[2:])
			if err != nil {
				fmt.Println("transfer Parse error")
				return 
			}
		case "addtoken" :
			err := addtokenFlag.Parse(os.Args[2:])
			if err != nil {
				fmt.Println("addtoken Parse error")
				return
			}
		case "sendtoken" : 
			err := sendtokenFlag.Parse(os.Args[2:])
			if err != nil {
				fmt.Println("sendtoken Parse error")
				return 
			}
		case "tokenbalance" :
			err := tokenbalanceFlag.Parse(os.Args[2:])
			if err != nil {
				fmt.Println("sendtoken Parse error")
				return 
			}
		default : 
			//输入错误，提供帮助文档
			c.Help()
			os.Exit(1)
	}
	//3. 根据解析的参数信息执行对应的功能
	if walletFlag.Parsed() {
		//创建账户
		fmt.Println(*walletAccount)
		if *walletAccount == "" {
			fmt.Println("walletAccount can not be null")
			return
		}

		//解决密码问题：让输入的密码不显示
		fmt.Println("请输入你的密码...")
		pass, _ := gopass.GetPasswd() ///这里的pass为切片类型
		c.CreateWallet(*walletAccount, string(pass))

	}

	if balanceFlag.Parsed() {
		fmt.Println(*walletBalanceAccount)
		if *walletBalanceAccount == "" {
			fmt.Println("walletBalanceAccount can not be null")
			return
		}

		//解决密码问题：让输入的密码不显示
		// fmt.Println("请输入你的密码...")
		// pass, _ := gopass.GetPasswd() ///这里的pass为切片类型

		c.GetBalance(*walletBalanceAccount)
	}

	if transferFlag.Parsed() {
		//实现转账
		fmt.Println(*transferName)
		fmt.Println(*transferToAddr)
		fmt.Println(*transferValue)
		if *transferName == "" || *transferToAddr == "" || *transferValue <= 0 {
			fmt.Println("fail to transfer: transfer Param error !!")
			c.Help()
			os.Exit(1)
		}
		//调用转账
		c.Transfer(*transferName, *transferToAddr, *transferValue)
	}

	if addtokenFlag.Parsed() {
		//实现token添加
		fmt.Println(*addtokenAddr)
		if *addtokenAddr == "" {
			fmt.Println("Failed to add token !")
			c.Help()
			os.Exit(1)
		}
		//调用Addtoken
		c.AddToken(*addtokenAddr)
	}

	if sendtokenFlag.Parsed() {
		//实现token转账
		fmt.Println(*sendtokenName)
		fmt.Println(*sendtokenSymbol)
		fmt.Println(*sendtokenToAddr)
		fmt.Println(*sendtokenAmount)
		if *sendtokenName == "" || *sendtokenSymbol == "" || *sendtokenToAddr == "" || *sendtokenAmount <= 0 {
			fmt.Println("Failed to send token ! ")
			c.Help()
			os.Exit(1)
		}
		c. SendToken(*sendtokenName, *sendtokenSymbol, *sendtokenToAddr, *sendtokenAmount)
	}

	if tokenbalanceFlag.Parsed() {
		fmt.Println(*tokenbalanceName)
		fmt.Println(*tokenbalanceSymbol)
		if *tokenbalanceName == "" || *tokenbalanceSymbol == "" {
			fmt.Println("Failed to get token balance! ")
			c.Help()
			os.Exit(1)
		}
		c.GetTokenBalance(*tokenbalanceName, *tokenbalanceSymbol)
	}
}


//参加创建账户功能  name = congcong, 这个name可以作为一个目录存储 即data/name/0x.... 
func (c *CLI) CreateWallet(name, pass string) {
		//1. 生成助记词
		mnemonic := NewMnemonic()
		//2. 通过助记词生成钱包     助记词--> 种子 --> 钱包 
		wallet, err := hd.NewFromMnemonic(mnemonic, pass)
		if err != nil {
			log.Panic("Failed to NewFromMnemonic", err)
		}
		//3. 通过钱包推导账户 
		path, _ := hd.ParseDerivationPath("m/44'/60'/0'/0/0")
		account, err := wallet.Derive(path, true)
		if err != nil {
			log.Panic("Failed to Derive", err)
		}
		fmt.Println(account.Address.Hex())
		//4. 通过账户得到私钥
		privateKey, err := wallet.PrivateKey(account)
		if err != nil {
			log.Panic("failed to get Private Key from account", err)
		}
		//5.通过私钥生成keystore对象  func (ks HDkeyStore) StoreKey(filename string, key *Key, auth string) error
		hdks := hdKeystore.NewHDkeyStore(c.DataPath+name , privateKey)
		hdks.StoreKey(hdks.JoinPath(account.Address.Hex()), &hdks.Key, pass)
		//用scrypt加密算法对privateKey进行加密，并校验私钥是否对应与私钥对应的地址
}

func (c *CLI) GetBalance(name string) {
	//1. 解析账户地址
	fromAddr := c.getAddr(name)
	if fromAddr == "" {
		fmt.Println("from address is null")
		return
	}

	//2. 连接到网络
	//func Dial(rawurl string) (*Client, error)
	client, err := rpc.Dial(c.NetWorkUrl)
	if err != nil {
		log.Panic("Failed to rpc dial, err", err)
	}
	defer client.Close()
	//3. 查询余额
	//func (c *Client) Call(result interface{}, method string, args ...interface{})
	var balance string
	err = client.Call(&balance, "eth_getBalance", fromAddr, "latest")
	if err != nil {
		log.Panic("Faied to get balance latest!")
	}
	value := hex2bigInt(balance)
	fmt.Println("balance = ", value.String())
	return
}



func (c *CLI) Transfer(transferName, toAddr string, value int64) {
	//1. 确认身份
	fromAddr := c.getAddr(transferName)
	if fromAddr == "" {
		log.Panic("fromAddr is null")
	}
	//2. 链接到以太坊网络
	//func Dial(rawurl string) (*Client, error)
	client, err := ethclient.Dial(c.NetWorkUrl)
	if err != nil {
		log.Panic("Failed to Dial", err)
	}
	defer client.Close()
	//3. 生成交易

	//nonce值与网络相关。。
	//生成nonce
	//func (ec *Client) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	nonce, err := client.NonceAt(context.Background(), common.HexToAddress(fromAddr), nil)
	if err != nil {
		log.Panic("failed to get nonce!!!")
	}
	amount := big.NewInt(value)
	gasLimit := uint64(300000)
	gasPrice := big.NewInt(21000000000)
	tx := types.NewTransaction(nonce, common.HexToAddress(toAddr), amount, gasLimit, gasPrice, []byte("salary"))
	/*
		func NewTransaction(nonce uint64, to common.Address, amount *big.Int,
							gasLimit uin64, gasPrice *big.Int, data []byte) *Transaction
	 */
	
	//4. 通过身份keystore进行签名
	//建立一个HDkeyStore对象
	//使用函数func NewHDkeyStore(path string, privateKey *ecdsa.PrivateKey) *HDkeyStore
	//解析文件获得key函数 : func (ks HDkeyStore) GetKey(addr common.Address, filename, auth string) (*keystore.Key, error)
	ks := hdKeystore.NewHDkeyStore(c.DataPath+transferName, nil)
	key, err := ks.GetKey(common.HexToAddress(fromAddr), ks.JoinPath(fromAddr), "123")
	if err != nil {
		log.Panic("failed to get key from keystore", err)
	}
	//can get privateKey from key (key.PrivateKey)
	ks.Key = *key ///很关键 如果没有 则key值为空
	//通过私钥对交易进行签名
	//func (ks HDkeyStore) SignTx(address common.Address, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error)
	signTx, err := ks.SignTx(common.HexToAddress(fromAddr), tx, nil) //nil可换成big.NewInt(20)即私链的network ID
	if err != nil {
		log.Panic("failed to sign tx")
	}

	//5. 发送交易
	//func (ec *client) SendTransaction(ctx context.Context, tx *types.Transaction) error
	err = client.SendTransaction(context.Background(), signTx) //这里的 signTx 是指签过名的tx
	if err != nil {
		log.Panic("client send transaction err", err)
	}
}

func (c *CLI) AddToken(addr string) {
	//1. 读取配置文件
	tokens := c.ReadToken()
	//2. 对配置文件进行校验
	if c.CheckToken(addr, tokens) {
		fmt.Println("token is exists", addr)
		return
	}
	//3. 连接到网络
	client, err := ethclient.Dial(c.NetWorkUrl)
	if err != nil {
		log.Panic("failed to ethclient Dial to Network URL")
	}
	defer client.Close()
	//4. 通过合约地址创建合约实例
	ins, err := erc20.NewErc20(common.HexToAddress(addr), client)
	if err != nil {
		log.Panic("Failed to create erc20")
	}
	sym, err := ins.Symbol(nil)
	if err != nil {
		log.Panic("failed to get symbol")
	}
	//5. 获得token的symbol
	tokens = append(tokens, TokenConfig{sym, addr})
	content, err := json.Marshal(tokens)
	if err != nil {
		log.Panic("Failed to json marshal tokens")
	}
	//6. 写入配置文件
	err = hdKeystore.WriteKeyFile(c.TokenFile, content)
	if err != nil {
		log.Panic("Failed to WriteKeyFile")
	}
}


	/*
		将文件解析放入到TokenConfig结构体中
	 */
func (c *CLI) ReadToken() []TokenConfig {
	tokens := []TokenConfig{}
	data, err := ioutil.ReadFile(c.TokenFile)
	if err != nil {
		log.Panic("Failed to ReadFile", err)
		//fmt.Println("Failed to ReadFile", err)
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &tokens)
		if err != nil {
			log.Panic("json unmarshal data err", err)
		}
	}
	return tokens
}

	/*
		如果 CheckToken 为真，则代表有重复的地址；
		如果 CheckToken 为假，则代表没有重复的地址，可添加。
	 */
func (c *CLI) CheckToken(addr string, tokens []TokenConfig) bool {
	for _, v := range tokens {
		if addr == v.Addr {
			return true
		}
	}
	return false 
}

func (c *CLI) SendToken(accname, symbol, toaddr string, amount int64) {
	// 1. 连接到网络
	client, err := ethclient.Dial(c.NetWorkUrl)
	if err != nil {
		log.Panic("failed to eth client dial")
	}
	defer client.Close()
	// 2. 生成合约实例
	//contract_addr ? 
	contract_addr := c.getContractAddr(symbol)
	if contract_addr == "" {
		fmt.Println("contract_addr is null, symbol not exist", symbol)
		return
	}
	ins, err := erc20.NewErc20(common.HexToAddress(contract_addr), client)
	if err != nil {
		log.Panic("Failed to new erc20")
	}
	// 3. 设置签名  
		// 1) NewHDkeyStore 
		  // func NewHDkeyStore(path string, privateKey *ecdsa.PrivateKey) *HDkeyStore
	HDks := hdKeystore.NewHDkeyStore(c.DataPath+accname, nil)
	fromAddr := c.getAddr(accname)
	key, err := HDks.GetKey(common.HexToAddress(fromAddr), HDks.JoinPath(fromAddr), "123")
	if err != nil {
		log.Panic("failed to get key from keystore", err)
	}
	HDks.Key = *key
		//func (ks HDkeyStore) NewTransactOpts() *bind.TransactOpts
	opts := HDks.NewTransactOpts()
	// 4. 合约调用
	value := big.NewInt(amount)
	//func (_Erc20 *Erc20Transactor) Transfer(opts *bind.TransactOpts, _to common.Address, _value *big.Int) (*types.Transaction, error)
	_, err = ins.Transfer(opts, common.HexToAddress(toaddr), value)//返回交易回执 可以在链上查看具体信息
	if err != nil {
		log.Panic("Failed to ins transfer", err)
	}
}

func (c *CLI) getContractAddr(symbol string) string {
	tokens := c.ReadToken() //得到[]TokenConfig  TokenConfig结构体中包含symbol 和 合约地址
	for _, v := range tokens {
		//rVal := reflect.ValueOf(v)
		if v.Symbol == symbol {
			return v.Addr
		}
	}
	return ""
}

func (c *CLI) GetTokenBalance(account, symbol string) {
	//1. 连接到网络
	client, err := ethclient.Dial(c.NetWorkUrl)
	if err != nil {
		log.Panic("failed to dial ethclient", err)
	}
	defer client.Close()
	//2. 生成合约实例
	contract_addr := c.getContractAddr(symbol)
	if contract_addr == "" {
		fmt.Println("contract_addr is null, symbol not exist", symbol)
		return
	}
	ins, err := erc20.NewErc20(common.HexToAddress(contract_addr), client)
	if err != nil {
		log.Panic("Failed to new erc20")
	}
	//3. 获得签名
	//HDks := hdKeystore.NewHDkeyStore(c.DataPath+account, nil)
	fromAddr := c.getAddr(account)
	// key, err := HDks.GetKey(common.HexToAddress(fromAddr), HDks.JoinPath(fromAddr), "123")
	// if err != nil {
	// 	log.Panic("failed to get key from keystore", err)
	// }
	// HDks.Key = *key
		// type CallOpts struct {
		// 	Pending     bool            // Whether to operate on the pending state or the last known one
		// 	From        common.Address  // Optional the sender address, otherwise the first account is used
		// 	BlockNumber *big.Int        // Optional the block number on which the call should be performed
		// 	Context     context.Context // Network context to support cancellation and timeouts (nil = no timeout)
		// }
	opts := &bind.CallOpts{
		From : common.HexToAddress(fromAddr),
	}
	//opts = *bind.CallOpts(opts)
	//4. 合约调用 
		//func (_Erc20 *Erc20Caller) BalanceOf(opts *bind.CallOpts, _owner common.Address) (*big.Int, error)
	addr := c.getAddr(account)
	amount, err := ins.BalanceOf(opts, common.HexToAddress(addr))
	if err != nil {
		fmt.Println("failed to get token balance, ", err)
		return
	}
	fmt.Println("balance = ", amount.String())
	return
}

func (c *CLI) getAddr(name string) string {
	infos, err := ioutil.ReadDir(c.DataPath + name)  //得到一个内容的切片类型
	if err != nil {
		log.Panic("Fail to read datapath")
	}
	for _, v := range infos {
		if !v.IsDir() {
			if strings.HasPrefix(v.Name(), "0x") {
				return v.Name()
			}
		}
	}
	return ""
}

//将16进制的数转换成10进制
func hex2bigInt(hex string) *big.Int {
	n := new(big.Int)
	n, _ = n.SetString(hex[2:], 16)//将16进制数转换成10进制 由于接受的数前面有0x故hex[2:]
	// if err != nil {
	// 	fmt.Println("hex SetString to 16 err", err)
	// 	return nil
	// }
	return n 
}

//生成助记词函数
func NewMnemonic() string{
	//1. NewEntropy 后面必须填上32的整数倍的数，并且在128-256之间
	entropy, _ := bip39.NewEntropy(128)
	//2. 助记词
	mnemonic, _ := bip39.NewMnemonic(entropy)
	//得到助记词
	fmt.Println(mnemonic)
	return mnemonic
}