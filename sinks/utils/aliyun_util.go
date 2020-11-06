package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AliyunContainerService/kube-eventer/sinks/sls"
	"github.com/denverdino/aliyungo/metadata"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog"
	"os"
	"strings"
	"time"
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

func ParseAKInfo() (*AKInfo, error) {
	m := metadata.NewMetaData(nil)
	var akInfo AKInfo
	if _, err := os.Stat(sls.ConfigPath); err == nil {
		//获取token config json
		encodeTokenCfg, err := ioutil.ReadFile(sls.ConfigPath)
		if err != nil {
			klog.Fatalf("failed to read token config, err: %v", err)
		}
		err = json.Unmarshal(encodeTokenCfg, &akInfo)
		if err != nil {
			klog.Fatalf("error unmarshal token config: %v", err)
		}
		keyring := akInfo.Keyring
		ak, err := Decrypt(akInfo.AccessKeyId, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode ak, err: %v", err)
		}

		sk, err := Decrypt(akInfo.AccessKeySecret, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode sk, err: %v", err)
		}

		token, err := Decrypt(akInfo.SecurityToken, []byte(keyring))
		if err != nil {
			klog.Fatalf("failed to decode token, err: %v", err)
		}
		layout := "2006-01-02T15:04:05Z"
		t, err := time.Parse(layout, akInfo.Expiration)
		if err != nil {
			klog.Errorf("failed to parse time layout, %v", err)
		}
		if t.Before(time.Now()) {
			klog.Error("invalid token which is expired")
		}
		klog.Info("get token by ram role.")
		akInfo.AccessKeyId = string(ak)
		akInfo.AccessKeySecret = string(sk)
		akInfo.SecurityToken = string(token)
	} else {
		roleName, err := m.RoleName()
		if err != nil {
			klog.Errorf("failed to get RoleName,because of %v", err)
			return nil, err
		}

		auth, err := m.RamRoleToken(roleName)
		if err != nil {
			klog.Errorf("failed to get RamRoleToken,because of %v", err)
			return nil, err
		}
		akInfo.AccessKeyId = auth.AccessKeyId
		akInfo.AccessKeySecret = auth.AccessKeySecret
		akInfo.SecurityToken = auth.SecurityToken
	}
	return &akInfo, nil
}

// Forked this method from knative eventing to support the same format of cloudevents subject in k8s world
// Creates a URI of the form found in object metadata selfLinks
// Format looks like: /apis/feeds.knative.dev/v1alpha1/namespaces/default/feeds/k8s-events-example
// KNOWN ISSUES:
// * ObjectReference.APIVersion has no version information (e.g. serving.knative.dev rather than serving.knative.dev/v1alpha1)
// * ObjectReference does not have enough information to create the pluaralized list type (e.g. "revisions" from kind: Revision)
//
// Track these issues at https://github.com/kubernetes/kubernetes/issues/66313
// We could possibly work around this by adding a lister for the resources referenced by these events.
func CreateSelfLink(o v1.ObjectReference) string {
	gvr, _ := meta.UnsafeGuessKindToResource(o.GroupVersionKind())
	versionNameHack := o.APIVersion

	// Core API types don't have a separate package name and only have a version string (e.g. /apis/v1/namespaces/default/pods/myPod)
	// To avoid weird looking strings like "v1/versionUnknown" we'll sniff for a "." in the version
	if strings.Contains(versionNameHack, ".") && !strings.Contains(versionNameHack, "/") {
		versionNameHack = versionNameHack + "/versionUnknown"
	}
	return fmt.Sprintf("/apis/%s/namespaces/%s/%s/%s", versionNameHack, o.Namespace, gvr.Resource, o.Name)
}
