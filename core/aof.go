package core

import (
	"fmt"
	"log"
	"maps"
	"os"
	"strings"

	"github.com/kepnok/bedis/config"
)

func dumpKey(fp *os.File, key string, obj *Obj) {
	cmd := fmt.Sprintf("SET %s %s", key, obj.Value)
	tokens := strings.Split(cmd, " ")
	fp.Write(Encode(tokens, false))
}

func DumpAllAOF() {
	fp, err := os.OpenFile(config.AOFFile, os.O_CREATE|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Print("error: ", err)
		return 
	}
	defer fp.Close()

	copyStore := make(map[string]*Obj)

	mu.Lock()
	maps.Copy(copyStore, store)
	mu.Unlock()

	log.Println("writing AOF file at ", config.AOFFile)
	for k, obj := range copyStore {
		dumpKey(fp, k, obj)
	}
	log.Println("AOF write complete")
}