package util

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

// GenerateRSAKey 生成RSA私钥和公钥，保存到文件中
func GenerateRSAKey(bits int) (err error, publicKeyText []byte, privateKeyText []byte) {
	//GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	//Reader是一个全局、共享的密码用强随机数生成器
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err, nil, nil
	}
	//保存私钥
	//通过x509标准将得到的ras私钥序列化为ASN.1 的 DER编码字符串
	X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
	var bufPrivateKey bytes.Buffer
	writerPrivateKey := bufio.NewWriter(&bufPrivateKey)
	pem.Encode(writerPrivateKey, &privateBlock)
	writerPrivateKey.Flush()
	privateKeyText = bufPrivateKey.Bytes()

	//保存公钥
	publicKey := privateKey.PublicKey
	//X509对公钥编码
	X509PublicKey, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		//logging.CLILog.Error(err)
		return err, nil, nil
	}
	var bufPublicKey bytes.Buffer
	writerPublicKey := bufio.NewWriter(&bufPublicKey)
	publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
	pem.Encode(writerPublicKey, &publicBlock)
	writerPublicKey.Flush()
	publicKeyText = bufPublicKey.Bytes()

	return nil, publicKeyText, privateKeyText
}

// GenerateRSAKeyFile 生成RSA私钥和公钥，保存到文件中
func GenerateRSAKeyFile(bits int, savedKeyFilePath string) (err error) {
	var publicKey, privateKey []byte
	err, publicKey, privateKey = GenerateRSAKey(bits)
	if err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(savedKeyFilePath, "public.pem"), publicKey, 0666); err != nil {
		return
	}
	if err = os.WriteFile(filepath.Join(savedKeyFilePath, "private.pem"), privateKey, 0666); err != nil {
		return
	}
	return nil
}

// RSADecryptFromPemText RSA解密
func RSADecryptFromPemText(cipherText []byte, pemFileText []byte) (plainText []byte, err error) {
	//pem解码
	block, _ := pem.Decode(pemFileText)
	//X509解码
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	//对密文进行解密
	plainText, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherText)
	//返回明文
	return
}
