package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func main() {
	mp := make(map[string]interface{})
	json.Unmarshal(bytes.Trim([]byte(`
	{
		"hello": "world",
		"just": {
			"sing": "a-song",
			"and": [1, 2, 3, 4]
		}
	}
	`), " \r\n\t"), &mp)
	fmt.Printf("%#v\r\n", mp)
}
