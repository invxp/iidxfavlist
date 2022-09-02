package iidxfavlist

import (
	"log"
)

//logf 打印日志(如果没有启用则打到控制台)
func (iidx *Iidxfavlist) logf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
