package util

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	aliyunECSRamURL      = "http://100.100.100.200/latest/meta-data/ram/security-credentials/"
	expirationTimeFormat = "2006-01-02T15:04:05Z"
)

type UpdateTokenFunction = func() (accessKeyID, accessKeySecret, securityToken string, expireTime time.Time, err error)

var errNoFile = errors.New("no secret file")

// AKInfo ...
type AKInfo struct {
	AccessKeyId     string `json:"access.key.id"`
	AccessKeySecret string `json:"access.key.secret"`
	SecurityToken   string `json:"security.token"`
	Expiration      string `json:"expiration"`
	Keyring         string `json:"keyring"`
}

// SecurityTokenResult ...
type SecurityTokenResult struct {
	AccessKeyId     string
	AccessKeySecret string
	Expiration      string
	SecurityToken   string
	Code            string
	LastUpdated     string
}

func getToken() (result []byte, err error) {
	client := http.Client{
		Timeout: time.Second * 3,
	}
	var respList *http.Response
	respList, err = client.Get(aliyunECSRamURL)
	if err != nil {
		return nil, err
	}
	defer respList.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(respList.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := string(body)
	bodyStr = strings.TrimSpace(bodyStr)
	roles := strings.Split(bodyStr, "\n")
	role := roles[0]

	var respGet *http.Response
	respGet, err = client.Get(aliyunECSRamURL + role)
	if err != nil {
		return nil, err
	}
	defer respGet.Body.Close()
	body, err = ioutil.ReadAll(respGet.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func decrypt(s string, keyring []byte) ([]byte, error) {
	cdata, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(keyring)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()

	iv := cdata[:blockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(cdata)-blockSize)

	blockMode.CryptBlocks(origData, cdata[blockSize:])

	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func getAKFromLocalFile(configFilePath string) (accessKeyID, accessKeySecret, securityToken string, expireTime time.Time, err error) {
	if _, err = os.Stat(configFilePath); err == nil {
		var akInfo AKInfo
		//获取token config json
		encodeTokenCfg, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}
		err = json.Unmarshal(encodeTokenCfg, &akInfo)
		if err != nil {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}
		keyring := akInfo.Keyring
		ak, err := decrypt(akInfo.AccessKeyId, []byte(keyring))
		if err != nil {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}

		sk, err := decrypt(akInfo.AccessKeySecret, []byte(keyring))
		if err != nil {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}

		token, err := decrypt(akInfo.SecurityToken, []byte(keyring))
		if err != nil {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}
		layout := "2006-01-02T15:04:05Z"
		t, err := time.Parse(layout, akInfo.Expiration)
		if err != nil {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}
		if t.Before(time.Now()) {
			err = errors.New("invalid token which is expired")
		}
		akInfo.AccessKeyId = string(ak)
		akInfo.AccessKeySecret = string(sk)
		akInfo.SecurityToken = string(token)

		if err != nil {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}
		return akInfo.AccessKeyId, akInfo.AccessKeySecret, akInfo.SecurityToken, t, nil
	}
	return accessKeyID, accessKeySecret, securityToken, expireTime, errNoFile
}

func updateTokenFunction(configFilePath string) (accessKeyID, accessKeySecret, securityToken string, expireTime time.Time, err error) {
	if configFilePath != "" {
		accessKeyID, accessKeySecret, securityToken, expireTime, err = getAKFromLocalFile(configFilePath)
		if err != errNoFile {
			return accessKeyID, accessKeySecret, securityToken, expireTime, err
		}
	}
	var tokenResultBuffer []byte
	for tryTime := 0; tryTime < 3; tryTime++ {
		tokenResultBuffer, err = getToken()
		if err != nil {
			continue
		}
		var tokenResult SecurityTokenResult
		err = json.Unmarshal(tokenResultBuffer, &tokenResult)
		if err != nil {
			continue
		}
		if strings.ToLower(tokenResult.Code) != "success" {
			tokenResult.AccessKeySecret = "x"
			tokenResult.SecurityToken = "x"
			continue
		}
		expireTime, err := time.Parse(expirationTimeFormat, tokenResult.Expiration)
		if err != nil {
			tokenResult.AccessKeySecret = "x"
			tokenResult.SecurityToken = "x"
			continue
		}
		return tokenResult.AccessKeyId, tokenResult.AccessKeySecret, tokenResult.SecurityToken, expireTime, nil
	}
	return accessKeyID, accessKeySecret, securityToken, expireTime, err
}

// NewTokenUpdateFunc create a token update function for ACK or ECS
func NewTokenUpdateFunc(role string, configFilePath string) (tokenUpdateFunc UpdateTokenFunction, shutdown chan struct{}) {
	return func() (accessKeyID string, accessKeySecret string, securityToken string, expireTime time.Time, err error) {
		return updateTokenFunction(configFilePath)
	}, make(chan struct{})
}
