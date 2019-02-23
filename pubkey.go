package nacos

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/dxvgef/go-lib/httplib"
)

// SaasPubKey 对象
type SaasPubKey struct {
	PublicKeyStr string
}

// GetPubKeyFromAuth 获取 public key
func (t *SaasPubKey) GetPubKeyFromAuth(reqURL, clientID, secret string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	req := httplib.Post(reqURL)
	req.Param("clientId", clientID)
	req.Param("secret", secret)

	resp, err := req.Bytes()
	if err != nil {
		return err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return err
	}

	// log.Println(result)

	pubKey, ok := result["data"].(string)
	if !ok {
		return errors.New("pubKey 断言错误")
	}

	if pubKey == "" {
		return errors.New("公钥为空")
	}

	t.PublicKeyStr = pubKey

	return nil
}

// GetTokenFromAuth 获取accesstoken
func GetTokenFromAuth(reqURL, clientID, secret string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	// 注册中心获取 token
	req := httplib.Post(reqURL)
	req.Param("clientId", clientID)
	req.Param("secret", secret)

	resp, err := req.Bytes()
	if err != nil {
		return "", err
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if result["error"] != nil {
		errMsg, ok := result["error"].(string)
		if !ok {
			return "", errors.New("error 断言失败")
		}
		return "", errors.New(errMsg)
	}

	token, ok := result["data"].(string)
	if !ok {
		return "", errors.New("token 断言失败")
	}

	return token, nil
}

// VerifySignByPublicKey 验证token
func (t *SaasPubKey) VerifySignByPublicKey(tokenStr string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	tokenSlice := strings.Split(tokenStr, ".")

	// 判断长度
	if len(tokenSlice) != 3 {
		return errors.New("token 格式不对")
	}

	// 验证签名
	pubKeyByte, err := base64.StdEncoding.DecodeString(t.PublicKeyStr)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubKeyByte)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	pubk, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("key 断言失败")
	}

	signByte, err := base64.RawURLEncoding.DecodeString(tokenSlice[2])
	if err != nil {
		log.Println(err.Error())
		return err
	}

	hash := sha256.New()
	hash.Write([]byte(tokenSlice[0] + "." + tokenSlice[1]))
	err = rsa.VerifyPKCS1v15(pubk, crypto.SHA256, hash.Sum(nil), signByte)
	if err != nil {
		return err
	}

	return nil
}

// Valid 验证签名，验证过期时间
func (t *SaasPubKey) Valid(tokenStr string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	tokenSlice := strings.Split(tokenStr, ".")

	// 判断长度
	if len(tokenSlice) != 3 {
		return errors.New("token 格式不对")
	}

	// 判断是否过期
	b, err := base64.RawStdEncoding.DecodeString(tokenSlice[1])
	if err != nil {
		return err
	}

	type tokenResult struct {
		Sub    string
		UserID string
		Exp    int64
		Name   string
	}
	var r tokenResult
	err = json.Unmarshal(b, &r)
	if err != nil {
		return err
	}

	if r.Exp < time.Now().Unix() {
		return errors.New("token 过期")
	}

	// 验证签名
	pubKeyByte, err := base64.StdEncoding.DecodeString(t.PublicKeyStr)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	pubKey, err := x509.ParsePKIXPublicKey(pubKeyByte)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	pubk, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("key 断言失败")
	}

	signByte, err := base64.RawURLEncoding.DecodeString(tokenSlice[2])
	if err != nil {
		log.Println(err.Error())
		return err
	}

	hash := sha256.New()
	hash.Write([]byte(tokenSlice[0] + "." + tokenSlice[1]))
	err = rsa.VerifyPKCS1v15(pubk, crypto.SHA256, hash.Sum(nil), signByte)
	if err != nil {
		return err
	}

	return nil
}
