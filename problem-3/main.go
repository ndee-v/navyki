package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	// BucketSize - размер корзины, можно регулировать ограничитель запросов
	BucketSize = 10
	// FillFrequency - частота пополнения корзины, можно регулировать ограничитель запросов
	FillFrequency = time.Second
)

func main() {
	count := 100
	ch := make(chan int, count)

	attempt := 0
	tokenChan := make(chan struct{}, BucketSize)

	// Горутина, которая пополняет корзину с токенами
	go func() {
		t := time.NewTicker(FillFrequency)
		for {
			select {
			case <-t.C:
				log.Println("[INFO] fill token bucket")
				attempt = BucketSize
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
