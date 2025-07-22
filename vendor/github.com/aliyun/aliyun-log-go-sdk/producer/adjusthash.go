package producer

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const zero32Str = "00000000000000000000000000000000"

func toHex(x byte) byte {
	if x < 10 {
		return '0' + x
	}
	return 'a' - 10 + x
}

func getBitValue(val byte, bits int) byte {
	if bits >= 16 {
		return val
	}
	if bits >= 8 {
		return val & (0xf - 1)
	}
	if bits >= 4 {
		return val & (0xf - 3)
	}
	if bits >= 2 {
		return val & (0xf - 7)
	}
	return val & (0xf - 15)
}

func AdjustHash(shardhash string, buckets int) (string, error) {
	h := md5.New()
	h.Write([]byte(shardhash))
	cipherStr := h.Sum(nil)

	var destBuf strings.Builder
	destBuf.Grow(32)
	i := 0

	for buckets > 0 && i < len(cipherStr) {
		if (i & 0x01) == 0 {
			destBuf.WriteByte(toHex(getBitValue(cipherStr[i>>1]>>4, buckets)))
		} else {
			destBuf.WriteByte(toHex(getBitValue(cipherStr[i>>1]&0xF, buckets)))
		}
		buckets = buckets >> 4
		i++
	}
	destBuf.WriteString(zero32Str[0 : 32-i])
	return destBuf.String(), nil
}

func AdjustHashOld(shardhash string, buckets int) (string, error) {
	res := Md5ToBin(ToMd5(shardhash))
	x, err := BitCount(buckets)
	if err != nil {
		return "", err
	}
	tt := res[0:x]
	tt = FillZero(tt, 8)
	base, _ := strconv.ParseInt(tt, 2, 10)
	yy := strconv.FormatInt(base, 16)
	return FillZero(yy, 32), nil

}

// smilar as java Integer.bitCount
func BitCount(buckets int) (int, error) {
	bin := strconv.FormatInt(int64(buckets), 2)
	if strings.Contains(bin[1:], "1") || buckets <= 0 {
		return -1, errors.New(fmt.Sprintf("buckets must be a power of 2, got %v,and The parameter "+
			"buckets must be greater than or equal to 1 and less than or equal to 256.", buckets))
	}
	return strings.Count(bin, "0"), nil
}

func ToMd5(name string) string {
	h := md5.New()
	h.Write([]byte(name))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func Md5ToBin(md5 string) string {
	bArr, _ := hex.DecodeString(md5)
	res := ""
	for _, b := range bArr {
		res = fmt.Sprintf("%s%.8b", res, b)
	}
	return res
}

func FillZero(x string, n int) string {
	length := n - (strings.Count(x, "") - 1)
	for i := 0; i < length; i++ {
		x = x + "0"
	}
	return x
}
