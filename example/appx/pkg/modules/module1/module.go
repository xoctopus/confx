package module1

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
			fmt.Printf("module1 initialize failed: %v\n", err)
			panic(err)
		}
		fmt.Println("module1 initialized")
	}
}
