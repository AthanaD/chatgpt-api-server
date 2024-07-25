package utility

import (
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// GetEsyncValue 获取当前时间的esync值，该值为当前时间戳向下取整到最近24小时整数倍的时间戳
func GetEsyncValue(ctx g.Ctx) (esyncValue string) {
	// 获取当前时间戳
	currentTimestamp := time.Now().Unix()
	// 计算最近的86400秒整数倍时间戳
	roundedTimestamp := currentTimestamp - (currentTimestamp % 86400)
	// 转换为字符串返回
	esyncValue = strconv.FormatInt(roundedTimestamp, 10)
	return
}
