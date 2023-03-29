package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	bfPool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer([]byte{})
		},
	}
)

var TimeLayoutStr = "2006-01-02 15:04:05" //格式化时间
var TimeLayoutDate = "2006-01-02"

// JoinInts format int slice like:n1,n2,n3.
func JoinInts(is []int) string {
	if len(is) == 0 {
		return ""
	}
	if len(is) == 1 {
		return strconv.Itoa(is[0])
	}
	buf := bfPool.Get().(*bytes.Buffer)
	for _, i := range is {
		buf.WriteString(strconv.Itoa(i))
		buf.WriteByte(',')
	}
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	s := buf.String()
	buf.Reset()
	bfPool.Put(buf)
	return s
}

// SplitInts split string into int slice.
func SplitInts(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}
	sArr := strings.Split(s, ",")
	res := make([]int, 0, len(sArr))
	for _, sc := range sArr {
		i, err := strconv.Atoi(sc)
		if err != nil {
			return nil, err
		}
		res = append(res, i)
	}
	return res, nil
}

// Add add
func Add(x, y float64, places int) float64 {
	a := big.NewFloat(x)
	b := big.NewFloat(y)
	c := new(big.Float).Add(a, b)
	d, _ := c.Float64()
	return Round(d, places)
}

// Sub sub
func Sub(x, y float64, places int) float64 {
	a := big.NewFloat(x)
	b := big.NewFloat(y)
	c := new(big.Float).Sub(a, b)
	d, _ := c.Float64()
	return Round(d, places)
}

// Mul mul
func Mul(x, y float64, places int) float64 {
	a := big.NewFloat(x)
	b := big.NewFloat(y)
	c := new(big.Float).Mul(a, b)
	d, _ := c.Float64()
	return Round(d, places)
}

// Div 精度除法
func Div(x, y float64, places int) float64 {
	a := big.NewFloat(x)
	b := big.NewFloat(y)
	c := new(big.Float).Quo(a, b)
	d, _ := c.Float64()
	return Round(d, places)
}

// Round round
func Round(val float64, places int) float64 {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= 0.5 {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow
}

//校验是否在数组内
func InArray(needle interface{}, haystack interface{}) bool {
	val := reflect.ValueOf(haystack)
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if reflect.DeepEqual(needle, val.Index(i).Interface()) {
				return true
			}
		}
	case reflect.Map:
		for _, k := range val.MapKeys() {
			if reflect.DeepEqual(needle, val.MapIndex(k).Interface()) {
				return true
			}
		}
	default:
		panic("haystack: haystack type muset be slice, array or map")
	}

	return false
}

//校验是浮点数还是数字
func CheckPriceFloatOrNumber(inter interface{}) string {
	ps := fmt.Sprintf("%v", inter)
	strSlice := strings.Split(ps, ".")
	if len(strSlice) <= 1 {
		return ps
	}
	str0 := strSlice[0]
	str1 := strSlice[1]
	nf, _ := strconv.ParseFloat(str1, 64)
	if nf > 0 {
		return strings.TrimRight(ps, "0")
	}
	return str0
}

//生成随机数字
func RandNumber(num int) string {
	s := ""
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < num; i++ {
		randNum := rand.Intn(9)
		s = fmt.Sprintf("%s%d", s, randNum)
	}
	return s
}

//生成md5加密
func Md5V(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//加密
func BcryptSalt(str string) string {
	password := []byte(str)
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		logrus.Println(err)
	}
	return string(hash)
}

//解密
func CompareSalt(hashed string, plain string) bool {
	byteHash := []byte(hashed)
	plainStr := []byte(plain)
	err := bcrypt.CompareHashAndPassword(byteHash, plainStr)
	if err != nil {
		logrus.Println(err)
		return false
	}
	return true
}

func HexToTen(s string) *big.Int {
	byteValue, err := hex.DecodeString(s)
	if err != nil {
		fmt.Println(err)
	}
	IntValue := new(big.Int).SetBytes(byteValue)
	return IntValue
}

//UniqueSlice 去除切片重复的元素
func UniqueSlice(arrs []int32) (newArrs []int32) {
	arrMaps := make(map[int32]int)
	for _, va := range arrs {
		if _, ok := arrMaps[va]; !ok {
			arrMaps[va] = 1
			newArrs = append(newArrs, va)
		} else {
			continue
		}
	}
	return
}

//Intersect 取切片交集
func Intersect(lists ...[]int32) []int32 {
	var inter []int32
	mp := make(map[int32]int)
	l := len(lists)
	// 特判 只传了0个或者1个切片的情况
	if l == 0 {
		return make([]int32, 0)
	}
	if l == 1 {
		for _, s := range lists[0] {
			if _, ok := mp[s]; !ok {
				mp[s] = 1
				inter = append(inter, s)
			}
		}
		return inter
	}
	// 一般情况
	// 先使用第一个切片构建map的键值对
	for _, s := range lists[0] {
		if _, ok := mp[s]; !ok {
			mp[s] = 1
		}
	}
	// 除去第一个和最后一个之外的list
	for _, list := range lists[1 : l-1] {
		for _, s := range list {
			if _, ok := mp[s]; ok {
				// 计数+1
				mp[s]++
			}
		}
	}
	for _, s := range lists[l-1] {
		if _, ok := mp[s]; ok {
			if mp[s] == l-1 {
				inter = append(inter, s)
			}
		}
	}
	return inter
}

//DateFt 查询前几天数量
func DateFt(d, n int) string {
	t := time.Now()
	if d == 1 { //1表示天
		t = t.AddDate(0, 0, -n)
	} else if d == 2 { //2表示月
		t = t.AddDate(0, -n, 0)
	}
	return t.Format("2006-01-02 00:00:00")
}
