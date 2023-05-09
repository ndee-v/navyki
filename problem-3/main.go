package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	count := 100
	ch := make(chan int, count)
	bucketSize := 2
	attempt := 0
	tokenChan := make(chan struct{}, bucketSize)

	// Горутина, которая пополняет корзину с токенами
	go func() {
		t := time.NewTicker(time.Second)
		for {
			select {
			case <-t.C:
				log.Println("[INFO] fill token bucket")
				attempt = bucketSize
				for attempt > 0 {
					attempt--
					select {
					case tokenChan <- struct{}{}:
					default:
						break
					}
				}
			}
		}
	}()

	var wg sync.WaitGroup

	// Горутина для чтения значений из канала
	go func() {
		for value := range ch {
			fmt.Println(value)
		}
	}()

	// Горутины для записи значений в канал
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-tokenChan
			ch <- RPCCall()
		}()
	}

	wg.Wait()
	close(ch)
}

func RPCCall() int {
	return rand.Int()
}
