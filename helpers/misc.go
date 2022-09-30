package helpers

import (
	"encoding/json"
	"fmt"
)

func DumpFormatted(val interface{}) {
	str, _ := json.MarshalIndent(val, "", "    ")
	fmt.Println(string(str))
}
