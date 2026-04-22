// buildAzul: program that builds a website



package main

import (
    "fmt"
    "os"
    "log"
    "strings"

	"github.com/goccy/go-yaml"
    util "github.com/prr123/utility/utilLib"
)

type azulStruct struct {
	Nav map[string]string `yaml:"nav"`
	Blogs []string `yaml:"blogs"`
}

func main() {

	numarg := len(os.Args)
    flags:=[]string{"dbg", "yaml", "domain"}

    useStr := "/yaml=<yamlfile> /domain=<domain> [/dbg]"
    helpStr := "azul build site program"

    if numarg > len(flags) +1 {
        fmt.Println("too many arguments in cl!")
        fmt.Println("usage: %s %s\n", os.Args[0], useStr)
        os.Exit(-1)
    }

    if numarg == 1 || (numarg > 1 && os.Args[1] == "help") {
        fmt.Printf("help: %s\n", helpStr)
        fmt.Printf("usage is: %s %s\n", os.Args[0], useStr)
        os.Exit(1)
    }

    flagMap, err := util.ParseFlags(os.Args, flags)
    if err != nil {log.Fatalf("util.ParseFlags: %v\n", err)}

    dbg:= false
    _, ok := flagMap["dbg"]
    if ok {dbg = true}

    yFil := ""
   	yval, ok := flagMap["yaml"]
    if !ok {
        log.Fatalf("error -- no yaml flag provided!\n")
    } else {
        if yval.(string) == "none" {log.Fatalf("error -- no yaml file name provided!\n")}
        yFil = yval.(string)
        idx := strings.IndexByte(yFil, '.')
        if idx > -1 {log.Fatalf("error -- yaml file <%s> has an extension!\n", yFil)}
    }

    domain := "test.imgerp.eu"
    dval, ok := flagMap["domain"]
    if !ok {
        log.Fatalf("error -- no domain flag provided!\n")
    } else {
        if dval.(string) == "none" {log.Fatalf("error -- no domain provided!\n")}
        domain = dval.(string)
    }

	yamlFilDir := "/home/peter/cloud/domains/" + domain + "/yaml/"
	_, err = os.Stat(yamlFilDir)
	if err != nil {log.Fatalf("error -- no yaml dir: %v\n", err)}

	yamlFilnam := yamlFilDir + yFil + ".yaml"
	yamlB, err := os.ReadFile(yamlFilnam)
	if err != nil {log.Fatalf("error -- cannot read yaml file: %v\n", err)}

	if dbg {fmt.Printf("yaml len: %d\n", len(yamlB))}

	navmap := make(map[string]string)
	bloglist := make([]string,0, 128)

	sit := azulStruct{Nav: navmap, Blogs:bloglist}

	err = yaml.Unmarshal(yamlB, &sit)
	if err != nil {log.Fatalf("error -- cannot decode yaml file: %v\n", err)}

	for k,v := range sit.Nav {
		fmt.Printf("  k: %s v: %s\n",k,v)
	}
	for i:=0; i< len(sit.Blogs); i++ {
		fmt.Printf(" --%d: %s\n",i+1, sit.Blogs[i])
	}
	fmt.Println("*** buildAzul success *****")
}
