package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"go.uber.org/zap/zapcore"
	"io"
	"lib_chaos/log"
	"lib_chaos/utils"
	"os"
)

var x = "eyJhZmZlY3RfdGltZSI6MTY3NTMwNDk4MSwiYXBwX2lkIjoxMDY3Mzg1LCJhc3NldHMiOltdLCJjb2lucyI6W3siY29pbl9uYW1lIjoiY29pbiIsImNvaW5fdmFsdWUiOjEwMDAwMH1dLCJjb250ZXh0IjoiaSdtIHNvIGhhbmRzb21lIiwiY3VycmVuY3kiOjAsImV4cGlyZV90aW1lIjoxNjc3ODk2OTgxLCJpdGVtcyI6W3siaXRlbV9jb3VudCI6MTAwMDAwLCJpdGVtX2lkIjoxfV0sInJlY3Zlcl91aWRzIjpbMTAwNDFdLCJzZW5kZXJfbmFtZSI6InNhbW8gZ3JvdXAiLCJzdnJfaWQiOjUsInN2cl9yZWdpb24iOiJnbG9iYWwiLCJ0aXRsZSI6InZpcCBnaWZ0IiwidHlwZSI6InNlbmRfbWFpbCJ9"

func main() {
	test()
	dec := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(x)))
	b, e := io.ReadAll(dec)
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Println(string(b))
}

func test() {
	log.InitLogger(zapcore.DebugLevel, os.Stdout)
	log.Info("xxx")
	utils.PanicWrap(func() {
		panic("must")
	})
	log.Flush()
}
