package unity

import (
	"io/ioutil"
	"log"
)

var STRINGS_DAT []byte

func init() {
	var err error
	if STRINGS_DAT, err = ioutil.ReadFile("strings.dat"); err != nil {
		log.Fatal(err)
	}
}
