开发一款基于Ethereum属于自己的钱包wallet
## 基础概念

**助记词**  

	可以理解为是简化版的私钥，为了记住自己的私钥而创建的。由12个英文单词组成
	
**私钥** 

	在区块链中，私钥是从椭圆曲线上得到的，相当于椭圆曲线上的一个点就对应一个私钥。所以私钥比较随机且不同的账户对应不同的私钥。私钥是代表着你的所有权，可以拥有操作你账户的所有权限。
	
**公钥**

	公钥是通过私钥推导而来的，私钥可以推导出公钥，但是公钥不能反推私钥。
	
**账户地址**

	账户地址是由公钥通过hash函数得到的。

**keystore文件 **

	keystore文件相当于为了方便使用而设计的，主要是由私钥生成的，内容为json格式的文本，而keystore文件只有通过个人的password才能解析出私钥。这个设计就十分符合我们现在生活中的使用场景。
	
**钱包**

	用于存储私钥的软件。

其中，HD钱包（分层确定性钱包），方便私钥的记忆，适合于集团化管理的钱包。

## 钱包的作用
* **签名**
> 钱包主要是用自己的私钥去给交易进行签名，使得交易变成签名交易，从而使得交易变得有效，不可以篡改以及不可抵赖。
同时需要保证签名的交易不可以被模仿，即签名交易不可模仿为创建另外一笔交易。
* **交易**
> 在ethereum中，钱包可以支持ether以及token交易，其中token是指通过智能合约产生的虚拟代币（以太坊独有的）。

	ether交易的流程：（API调用）
		生成一笔交易 --> 对交易进行签名 --> 连接到eth网络中 --> 对交易进行广播
token交易的流程：（合约调用）
		智能合约上发行部署token --> 连接节点 --> 对合约进行实例化 --> 对身份进行验证 --> 调用所需要发起的 交易

### 创建钱包
* 首先，创建一款钱包需要创建的东西有：
> 助记词，私钥，keystore文件，账户地址

* 助记词生成
1. 生成Entropy（128bit的数据），加上4bit的一个哈希转换的校验位；

2. 将132bit的数据以每11bit一份，切分成12等份；

3. 通过查字典，每一份都会对应一个单词，最终得到助记词。

代码定义实现：（这里使用的是bip39这套议案）
```
func NewMnemonic() string{
	//1. 生成Entropy
	entropy, _ := bip39.NewEntropy(128)
	//2. 助记词
	mnemonic, _ := bip39.NewMnemonic(entropy)
	//得到助记词
	fmt.Println(mnemonic)
	return mnemonic
}
```
其中，NewEntropy 后面必须填上32的整数倍的数，并且在128-256之间。
执行后可以得到生成的助记词为：（每次执行的助记词是不同的，这里的助记词就是个人的私钥缩略版，要妥善保管）
> soccer legal measure all limb wood obey skill belt problem unusual donor

	生成的助记词可以进行验证，通过以太坊提供的源码，可以尝试下是否可以推到出对应的账户。
```
wallet, err := hd.NewFromMnemonic(mnemonic, "password")
if err != nil {
	log.Panic("Failed to NewFromMnemonic", err)
}
path, _ := hd.ParseDerivationPath("m/44'/60'/0'/0/0")
account, err := wallet.Derive(path, true)
if err != nil {
	log.Panic("Failed to Derive", err)
}
fmt.Println(account.Address.Hex())
```
可以通过上述给出的助记词推导出以太坊账户，即验证上述助记词推导为正确的，结果如下：
> soccer legal measure all limb wood obey skill belt problem unusual donor
0x071a08F584c1abD3eE040D92860CAAf4aABaE10d

其中,`hd.ParseDerivationPath`函数中需要传入的是一个root推导路径，格式为：
> m / purpose' / coin_type' / account' / change / address_index；

其中 bip44 的提案`purpose = 44`,`purpose`代表的是提案；`coin_type`代表的是币种。ethereum默认的是`m/44'/60'/0'/0 + 账户索引`。

#### HD钱包
为什么需要HD钱包？
由于一个人可能会有一堆私钥，私钥多可以更好地保护个人的隐私。

##### BIP32/EIP32
* 为了满足个人多私钥的需求，根据BIP32提议：
> 根据一个随机数种子，通过分层确定性推导的方式来得到多个私钥，这样保存的时候，只需要保存一个种子即可，私钥可以通过种子推导出来。

* 分层钱包的结构如下：
seed(种子) --> 主私钥（一个） --> 父私钥（多个） --> 子私钥（多个） --> 孙私钥（多个）
**解释描述**：一个种子可以推导出一个主秘钥，每个人有且只有一个主秘钥，可以根据主秘钥推导出多个父秘钥，每个父秘钥可以推导多个子秘钥，这样一层一层往下拓展，就可以满足多个秘钥的需求。通过将不同的秘钥用于不同的场景，有效保护了个人隐私。

* 推导主私钥过程：
密码学上安全的伪随机发生器 --> 根种子（128,256,512 bits） --> 循环进行HMAC-SHA512函数(单向哈希) --> 512bits输出
**解释描述**：通过伪随机发生器，区块链中大多是用椭圆曲线密码学算法，相当于在椭圆曲线取一个点作为随机数.通过随机数生成种子，得到一个128bits的倍数的种子以后，通过多次循环进行HMAC-SHA512算法，可以得出一个512bits的结果。得到的最终结果，左256bits作为`主私钥`；右256bits作为`主链编码`。

* 父私钥推导子私钥过程：
分为两种方法：
1）通过父私钥+索引号进行推导；
2）通过父公钥+父链编码+索引号进行推导。
要点：
	+ 父私钥可以推导得出父公钥
	+ 第一种方法中：索引号为2^31~2^32
	  第二种方法中：索引号小于2^31
	+ 第二种方法推导中，如果缺少父链编码不可以推导出子私钥。
	+ 所有的推导过程均为单向推导，即子私钥不能推导父私钥，也不能推导兄弟私钥。
	
##### BIP39/EIP39
上述BIP32中是通过一个根种子的方式来生成主私钥，从而后续推导出大量的子孙私钥。但是在实际操作中，随机种子相对是一个毫无规则的长串数字，十分不方便记忆。因此，推出BIP39/EIP39提案。
BIP39/EIP39提案主要是**以助记词的方式**来生成种子。
* 助记词的生成：
	生成一个128位的随机数 + 4位的校验位；
	以每11位进行划分，划分位12个二进制数；
	进而通过hash字典得到12个单词。
	
得到助记词后，可以通过助记词配合秘要拉伸函数（常用的PBKDF2）将其转化为种子。
* 以助记词+可选密码的方式，通过随机函数（HMAC函数）进行多次重复运算，最终推导出512位的密钥种子，从而通过种子来构建确定性钱包并派生后续的分层私钥。
其中，可选密码可以提高暴力破解的难度。

代码实现：

首先需要用到的包：
`"github.com/tyler-smith/go-bip39"`
`"github.com/btcsuite/btcd/chaincfg"`
`"github.com/btcsuite/btcutil/hdkeychain"`

代码如下：
```go
//1. 生成助记词
entropy, _ := bip39.NewEntropy(128) //128位随机数
mnemonic, _ := bip39.NewMnemonic(entropy)
//2.通过助记词和密码生成种子
seed := bip39.NewSeed(mnemonic, "pwd")
fmt.Println(seed)
//3. 通过种子生成主秘钥
masterKey, _ := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
fmt.Println(masterKey)
//4. 通过种子构建钱包
hdWallet, _ := hd.NewFromSeed(seed)
```
##### BIP44/EIP44
BIP44/EIP44提案将秘钥推导的原模型变为一个秘钥路径，用‘/’隔开；
五个预定义树状层级结构：
m'/purpose'/coin'/account'/change'/address_index

其中，
* m为固定的；
* purpose代表的是提案，例如BIP44/EIP44 对应44
* coin代表的是币种，例如0：比特币，1：测试币，60：以太币
* account代表的是所持有币对应的账户索引
* change有两种状态，0：外部（用于接收）；1：内部（用于返回交易变更）
* address_index代表的是地址索引，一般不超过20
代码如下：
```go
path, _ := hd.ParseDerivationPath("m/44'/60'/0'/0/0")
account, _ := hdWallet.Derive(path, true)
hdWallet.PublicKey(account)
hdWallet.PrivateKey(account)
```
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	
	

