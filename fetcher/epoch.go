package fetcher

import (
	"fmt"
	"time"
)

func FetchEpoch() {
	now := time.Now().Unix()
	fmt.Printf("--- UNIX Epoch ---\n%d\n\n", now)
}
