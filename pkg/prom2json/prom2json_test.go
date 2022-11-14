package prom2json

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestProm2json(t *testing.T) {
	bt := GetProm2JsonStruct("http://127.0.0.1:9100/metrics")
	bty, _ := json.Marshal(bt)
	fmt.Printf("result:%v\n", string(bty))

}
