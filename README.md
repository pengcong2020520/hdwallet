# wallet
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






















