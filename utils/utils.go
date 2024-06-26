package utils

import (
	"fmt"
	"net"
	"runtime"
	"strings"
)

// 显示错误时运行时的堆栈信息
func Trace(errMessage string) string {
	var pcstack [32]uintptr
	// Callers 将调用函数的返回程序计数器填入调用 goroutine 堆栈的片段 pc。
	// 参数 skip 是在 pc 中记录之前要跳过的堆栈帧数，0 表示 Callers 本身的帧，1 表示 Callers 的调用者。它返回写入 pc 的条目数。
	n := runtime.Callers(3, pcstack[:])

	// Using Builder optimize speed.
	var str strings.Builder
	str.WriteString(errMessage + "\nTraceback:")
	for _, pc := range pcstack[:n] {
		function := runtime.FuncForPC(pc)
		file, line := function.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

// IsValidPerrAddr 判断是否满足 ip:port 的格式
func IsValidPerrAddr(addr string) bool {
	token1 := strings.Split(addr, ":")
	if len(token1) != 2 {
		return false
	}
	// 判断token1[0]是否为y一个ip地址
	ip := net.ParseIP(token1[0])
	if token1[0] == "localhost" || ip != nil {
		return true
	}
	return false
}
