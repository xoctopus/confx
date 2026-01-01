package module3

import (
	"context"
	"fmt"
	"time"
)

func Init(ctx context.Context) error {
	time.Sleep(time.Second)
	return nil
}

func InitRunner(ctx context.Context) func() {
	return func() {
		if err := Init(ctx); err != nil {
			fmt.Printf("module3 initialize failed: %v\n", err)
			panic(err)
		}
		fmt.Printf("module3 initialized\n")
	}
}
