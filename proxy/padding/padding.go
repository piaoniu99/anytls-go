package padding

import (
	"anytls/util"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/sagernet/sing/common/atomic"
)

var defaultPaddingScheme = []byte(`stop=8
0=34-120
1=100-400
2=400-500,c,500-1000,c,400-500,c,500-1000,c,500-1000,c,400-500
3=500-1000
4=500-1000
5=500-1000
6=500-1000
7=500-1000`)

var paddingScheme atomic.TypedValue[util.StringMap]
var PaddingSchemeRaw atomic.TypedValue[[]byte]
var PaddingStop atomic.Uint32
var PaddingMd5 atomic.TypedValue[string]

const CheckMark = -1

func init() {
	UpdatePaddingScheme(defaultPaddingScheme)
}

func UpdatePaddingScheme(b []byte) bool {
	scheme := util.StringMapFromBytes(b)
	if len(scheme) == 0 {
		return false
	}
	if stop, err := strconv.Atoi(scheme["stop"]); err == nil {
		PaddingStop.Store(uint32(stop))
	} else {
		return false
	}
	PaddingSchemeRaw.Store(b)
	paddingScheme.Store(scheme)
	PaddingMd5.Store(fmt.Sprintf("%x", md5.Sum(b)))
	return true
}

func GenerateRecordPayloadSizes(pkt uint32) (pktSizes []int) {
	scheme := paddingScheme.Load()
	if s, ok := scheme[strconv.Itoa(int(pkt))]; ok {
		sRanges := strings.Split(s, ",")
		for _, sRange := range sRanges {
			sRangeMinMax := strings.Split(sRange, "-")
			if len(sRangeMinMax) == 2 {
				_min, err := strconv.ParseInt(sRangeMinMax[0], 10, 64)
				if err != nil {
					continue
				}
				_max, err := strconv.ParseInt(sRangeMinMax[1], 10, 64)
				if err != nil {
					continue
				}
				_min, _max = min(_min, _max), max(_min, _max)
				i, _ := rand.Int(rand.Reader, big.NewInt(_max-_min))
				pktSizes = append(pktSizes, int(i.Int64()+_min))
			} else if sRange == "c" {
				pktSizes = append(pktSizes, CheckMark)
			}
		}
	}
	return
}
