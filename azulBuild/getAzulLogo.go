package main

import (
	"fmt"
	"log"
	"os"
	"bytes"
)

func main() {

	logoFilnam := "/home/peter/cloud/domains/test.imgerp.eu/html/indexTest.html"
	logBt, err := os.ReadFile(logoFilnam)
	if err !=nil {log.Fatalf("error -- cound not find index file! %v\n", err)}

	fmt.Printf("info -- found ndexTest File: %d!\n", len(logBt))

	idx := bytes.Index(logBt, []byte(";base64"))
	if idx == -1 {log.Fatalf("error -- could not find base64\n")}

	fmt.Printf("hd: %s\n", logBt[:idx+8])

	idx2 := bytes.IndexByte(logBt[idx+7:], '\n')
	if idx2 == -1 {log.Fatalf("error -- could not find base64 ret\n")}

	logtxt := logBt[idx +8: idx+4+idx2]
	fmt.Printf("idx+7: %d, idx2: %d\n%s\n",idx, idx2, logtxt)

	logoB64Filnam := "/home/peter/cloud/domains/test.imgerp.eu/img/fav.b64"
	err = os.WriteFile(logoB64Filnam, logtxt, 0666)
	if err !=nil {log.Fatalf("error -- cound not write logo file! %v\n", err)}

	fmt.Println("**** success ****")
}
