package main

import (
	"fmt"
	"time"
)

func main() {
	tsArr := []int64{1640661475,
		1640661488,
		1640661501,
		1640661944,
		1640662427,
		1640662700,
		1640662999,
		1640663223,
		1640663336,
		1640663894}
	for _, ts := range tsArr {
		parsedTime := time.Unix(ts, 0)
		fmt.Println(parsedTime.Unix(), parsedTime.Format("2006-01-02 15:04:05Z07:00"))
	}
	//fmt.Println(hex.EncodeToString([]byte("abcdef-\\+")))
	//fired := false
	//timer := time.NewTimer(time.Second)
	//defer func() {
	//	if !fired {
	//		if !timer.Stop() {
	//			<-timer.C
	//		}
	//	}
	//}()
}
