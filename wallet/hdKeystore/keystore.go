package hdKeystore

import (
	"fmt"
	"os"
	"io/ioutil"
	"path/filepath"
	_"io"
	"crypto/ecdsa"
	"math/big"


	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	_"github.com/ethereum/go-ethereum/accounts"

	"go_code/hdwallet/utils"

)

// CallOpts is the collection of options to fine tune a contract call request.
// type CallOpts struct {
// 	Pending     bool            // Whether to operate on the pending state or the last known one
// 	From        common.Address  // Optional the sender address, otherwise the first account is used
// 	BlockNumber *big.Int        // Optional the block number on which the call should be performed
// 	Context     context.Context // Network context to support cancellation and timeouts (nil = no timeout)
// }

type HDkeyStore struct {
	keysDirPath string
	scryptN     int
	scryptP     int
	Key         keystore.Key
	//mu          sync.RWMutex
	// skipKeyFileVerification disables the security-feature which does
	// reads and decrypts any newly created keyfiles. This should be 'false' in all
	// cases except tests -- setting this to 'true' is not recommended.
	//skipKeyFileVerification bool
}

func NewHDkeyStore(path string, privateKey *ecdsa.PrivateKey) *HDkeyStore {
	//生成Key
	uuid := []byte(utils.NewRandom())
	if privateKey == nil {
		return &HDkeyStore{
			keysDirPath : path,
			scryptN : keystore.LightScryptN,  //Scrypt加密算法中的N参数
			scryptP : keystore.LightScryptP,  //Scrypt加密算法中的P参数
			Key : keystore.Key{},
		}
	}
	key := keystore.Key{
		Id : uuid, // 通过随机生成
		Address : crypto.PubkeyToAddress(privateKey.PublicKey),// 通过公钥生成
		PrivateKey : privateKey,
	}
	//LightScryptN is the N parameter of Scrypt encryption algorithm, 
	//using 4MB memory and taking approximately 100ms CPU time on a modern processor.
	return &HDkeyStore{
		keysDirPath : path,
		scryptN : keystore.LightScryptN,  //Scrypt加密算法中的N参数
		scryptP : keystore.LightScryptP,  //Scrypt加密算法中的P参数
		Key : key,
	}

}

func (ks HDkeyStore) GetKey(addr common.Address, filename, auth string) (*keystore.Key, error) {
	// Load the key from the keystore and decrypt its contents
	keyjson, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	key, err := keystore.DecryptKey(keyjson, auth)
	if err != nil {
		return nil, err
	}
	// Make sure we're really operating on the requested key (no swap attacks)
	if key.Address != addr {
		return nil, fmt.Errorf("key content mismatch: have account %x, want %x", key.Address, addr)
	}
	return key, nil
}

// StoreKey generates a key, encrypts with 'auth' and stores in the given directory


func (ks HDkeyStore) StoreKey(filename string, key *keystore.Key, auth string) error {
	keyjson, err := keystore.EncryptKey(key, auth, ks.scryptN, ks.scryptP)
	if err != nil {
		return err
	}
	// Write into temporary file
	tmpName, err := writeTemporaryKeyFile(filename, keyjson)
	if err != nil {
		return err
	}
	// if !ks.skipKeyFileVerification {
	// 	// Verify that we can decrypt the file with the given password.
	// 	_, err = ks.GetKey(key.Address, tmpName, auth)
	// 	if err != nil {
	// 		msg := "An error was encountered when saving and verifying the keystore file. \n" +
	// 			"This indicates that the keystore is corrupted. \n" +
	// 			"The corrupted file is stored at \n%v\n" +
	// 			"Please file a ticket at:\n\n" +
	// 			"https://github.com/ethereum/go-ethereum/issues." +
	// 			"The error was : %s"
	// 		//lint:ignore ST1005 This is a message for the user
	// 		return fmt.Errorf(msg, tmpName, err)
	// 	}
	// }
	return os.Rename(tmpName, filename)
}

func (ks HDkeyStore) JoinPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join(ks.keysDirPath, filename)
}

func (ks HDkeyStore) SignTx(address common.Address, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	
	//sign the transaction and verify the sender to avoid hardware fault surprise
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, ks.Key.PrivateKey)
	if err != nil {
		return nil, err 
	}

	//remix->verify
	msg, err := signedTx.AsMessage(types.HomesteadSigner{})
	if err != nil {
		return nil, err 
	}
	sender := msg.From()
	if sender != address {
		return nil, fmt.Errorf("signer mismatch: expected")
	}

	return signedTx, nil
}


func WriteKeyFile(file string, content []byte) error {
	name, err := writeTemporaryKeyFile(file, content)
	if err != nil {
		return err
	}
	return os.Rename(name, file)
}

func writeTemporaryKeyFile(file string, content []byte) (string, error) {
	// Create the keystore directory with appropriate permissions
	// in case it is not present yet.
	const dirPerm = 0700
	if err := os.MkdirAll(filepath.Dir(file), dirPerm); err != nil {
		return "", err
	}
	// Atomic write: create a temporary hidden file first
	// then move it into place. TempFile assigns mode 0600.
	f, err := ioutil.TempFile(filepath.Dir(file), "."+filepath.Base(file)+".tmp")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil
}

// zeroKey zeroes a private key in memory.
func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

// func newKey(rand io.Reader) (*keystore.Key, error) {
// 	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return newKeyFromECDSA(privateKeyECDSA), nil
// }

func NewKeyFromECDSA(privateKeyECDSA *ecdsa.PrivateKey) *keystore.Key {
	id := utils.NewRandom()
	key := &keystore.Key{
		Id:         []byte(id),
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}
	return key
}

// NewKeyedTransactor is a utility method to easily create a transaction signer
// from a single private key.
	//func NewKeyedTransactor(key *ecdsa.PrivateKey) *bind.TransactOpts

func (ks HDkeyStore) NewTransactOpts() *bind.TransactOpts {
	return bind.NewKeyedTransactor(ks.Key.PrivateKey)
}

