package iidxfavlist

import (
	"log"
)

//logfn 打印日志(如果没有启用则打到控制台)
func (iidx *Iidxfavlist) logfn(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

func (iidx *Iidxfavlist) logf(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

func (iidx *Iidxfavlist) panic(v ...interface{}) {
	log.Panic(v...)
}
