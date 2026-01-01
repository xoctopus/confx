package module2

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
			fmt.Printf("module2 initialize failed: %v\n", err)
			panic(err)
		}
		fmt.Println("module2 initialized")
	}
}
