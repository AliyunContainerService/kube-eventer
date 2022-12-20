package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/denverdino/aliyungo/metadata"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"time"
)

const (
	SLSConfigPath      = "/var/sls-addon/token-config"
	CMSConfigPath      = "/var/cms-addon/token-config"
	StsTokenTimeLayout = "2006-01-02T15:04:05Z"
)

type AKInfo struct {
	AccessKeyId     string `json:"access.key.id"`
	AccessKeySecret string `json:"access.key.secret"`
	SecurityToken   string `json:"security.token"`
	Expiration      string `json:"expiration"`
	Keyring         string `json:"keyring"`
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func Decrypt(s string, keyring []byte) ([]byte, error) {
	cdata, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		klog.Errorf("failed to decode base64 string, err: %v", err)
		return nil, err
	}
	block, err := aes.NewCipher(keyring)
	if err != nil {
		klog.Errorf("failed to new cipher, err: %v", err)
		return nil, err
	}
	blockSize := block.BlockSize()

	iv := cdata[:blockSize]
	blockMode := cipher.NewCBCDecrypter(block, iv)
	origData := make([]byte, len(cdata)-blockSize)

	blockMode.CryptBlocks(origData, cdata[blockSize:])

	origData = PKCS5UnPadding(origData)
	return origData, nil
}

func GetRegionFromEnv() (region string, err error) {
	region = os.Getenv("RegionId")
	if region == "" {
		return "", errors.New("not found region info in env")
	}
	return region, nil
}

func GetOwnerAccountFromEnv() (accountId string, err error) {
	accountId = os.Getenv("OwnerAccountId")
	if accountId == "" {
		return "", errors.New("not found owner account info in env")
	}
	return accountId, nil
}

func ParseRegion() (string, error) {
	region, err := GetRegionFromEnv()
	if err != nil {
		m := metadata.NewMetaData(nil)
		region, err = m.Region()
		if err != nil {
			klog.Errorf("failed to get Region, because of %v", err)
			return "", err
		}
	}
	return region, nil
}

func ParseRegionFromMeta() (string, error) {
	m := metadata.NewMetaData(nil)
	region, err := m.Region()
	if err != nil {
		klog.Errorf("failed to get Region, because of %v", err)
		return "", err
	}
	return region, nil
}

func ParseOwnerAccountId() (string, error) {
	accountId, err := GetOwnerAccountFromEnv()
	if err != nil {
		m := metadata.NewMetaData(nil)
		accountId, err = m.OwnerAccountID()
		if err != nil {
			klog.Errorf("failed to get OwnerAccount, because of %v", err)
			return "", err
		}
	}
	return accountId, nil
}

// GetAkInfo aliyun akInfo parse logic.
// include: 1. parse akInfo from configmap, 2. parse akInfo from metadata
func GetAkInfo(configPath string) *AKInfo {
	m := metadata.NewMetaData(nil)

	var akInfo AKInfo
	if _, err := os.Stat(configPath); err == nil {
		//获取token config json
		encodeTokenCfg, err := ioutil.ReadFile(configPath)
		if err != nil {
			klog.Fatalf("failed to read token config, configPath: %v,  err: %v", configPath, err)
		}
		err = json.Unmarshal(encodeTokenCfg, &akInfo)
		if err != nil {
			klog.Fatalf("error unmarshal token config, , configPath: %v,  err: %v", configPath, err)
		}
		keyring := akInfo.Keyring
		ak, err := Decrypt(akInfo.AccessKeyId, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode ak, configPath: %v,  err: %v", configPath, err)
		}

		sk, err := Decrypt(akInfo.AccessKeySecret, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode sk, configPath: %v,  err: %v", configPath, err)
		}

		token, err := Decrypt(akInfo.SecurityToken, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode token, configPath: %v,  err: %v", configPath, err)
		}
		layout := "2006-01-02T15:04:05Z"
		t, err := time.Parse(layout, akInfo.Expiration)
		if err != nil {
			klog.Errorf(err.Error())
		}
		if t.Before(time.Now()) {
			klog.Errorf("invalid token which is expired")
		}
		akInfo.AccessKeyId = string(ak)
		akInfo.AccessKeySecret = string(sk)
		akInfo.SecurityToken = string(token)
	} else {
		klog.Info("get token from metadata.")
		var (
			rolename string
			err      error
		)
		if rolename, err = m.RoleName(); err != nil {
			klog.Errorf("Failed to refresh sts rolename, configPath: %v, because of %s\n", configPath, err.Error())
			return nil
		}

		role, err := m.RamRoleToken(rolename)

		if err != nil {
			klog.Errorf("Failed to refresh sts token, configPath: %v, because of %s\n", configPath, err.Error())
			return nil
		}
		akInfo.AccessKeyId = role.AccessKeyId
		akInfo.AccessKeySecret = role.AccessKeySecret
		akInfo.SecurityToken = role.SecurityToken
	}
	return &akInfo
}

//func ParseAKInfoFromConfigPath() (*AKInfo, error) {
//	var akInfo AKInfo
//	var err error
//	if _, err = os.Stat(ConfigPath); err == nil {
//		//获取token config json
//		encodeTokenCfg, err := ioutil.ReadFile(ConfigPath)
//		if err != nil {
//			klog.Fatalf("failed to read token config, err: %v", err)
//		}
//		err = json.Unmarshal(encodeTokenCfg, &akInfo)
//		if err != nil {
//			klog.Fatalf("error unmarshal token config: %v", err)
//		}
//		keyring := akInfo.Keyring
//		ak, err := Decrypt(akInfo.AccessKeyId, []byte(keyring))
//		if err != nil {
//			klog.Fatalf("failed to decode ak, err: %v", err)
//		}
//
//		sk, err := Decrypt(akInfo.AccessKeySecret, []byte(keyring))
//		if err != nil {
//			klog.Fatalf("failed to decode sk, err: %v", err)
//		}
//
//		token, err := Decrypt(akInfo.SecurityToken, []byte(keyring))
//		if err != nil {
//			klog.Fatalf("failed to decode token, err: %v", err)
//		}
//
//		t, err := time.Parse(StsTokenTimeLayout, akInfo.Expiration)
//		if err != nil {
//			klog.Errorf("failed to parse time layout, %v", err)
//		}
//		if t.Before(time.Now()) {
//			klog.Error("invalid token which is expired")
//		}
//		klog.Info("get token by ram role.")
//		akInfo.AccessKeyId = string(ak)
//		akInfo.AccessKeySecret = string(sk)
//		akInfo.SecurityToken = string(token)
//
//		return &akInfo, nil
//	}
//
//	return nil, err
//}
