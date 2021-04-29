package main

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

var th = 6

func main() {

	print(rand.Int31n(156))
}

func ExecutePipeline(jobs ...job) {
	wg := sync.WaitGroup{}
	in := make(chan interface{})
	for _, j := range jobs {
		out := make(chan  interface{})
		wg.Add(1)

		go func(j job, in,out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			j(in, out)
		}(j, in, out, &wg)

		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	for i := range in {
		wg.Add(1)
		go func(i interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
			wgData := &sync.WaitGroup{}
			data := strconv.Itoa(i.(int))
			var crc32Data, crc32Md5Data string

			mu.Lock()
			md5Data := DataSignerMd5(data)
			mu.Unlock()

			wgData.Add(2)
			defer wg.Done()
			go func() {
				defer wgData.Done()
				crc32Data = DataSignerCrc32(data)
			}()
			go func() {
				defer wgData.Done()
				crc32Md5Data = DataSignerCrc32(md5Data)
			}()
			wgData.Wait()

			out <- crc32Data + "~" + crc32Md5Data
		}(i, &wg, &mu)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := sync.WaitGroup{}
	for i := range in {
		wg.Add(1)

		go func(i interface{}, wg *sync.WaitGroup) {
			defer wg.Done()
			wgData := &sync.WaitGroup{}
			muData := &sync.Mutex{}
			results := make([]string, th)

			for t := 0; t < th; t++ {
				wgData.Add(1)
				// Make data for each th=(0..5)
				data := strconv.Itoa(t) + i.(string)

				go func(i int) {
					defer wgData.Done()
					crc32Result := DataSignerCrc32(data)

					muData.Lock()
					results[i] = crc32Result
					muData.Unlock()
				}(t)
			}
			wgData.Wait()

			// Make results as one result var
			var result string
			for _, r := range results {result += r}

			out <- result
		}(i, &wg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for i := range in {
		results = append(results, i.(string))
	}
	sort.Strings(results)

	out <- strings.Join(results, "_")
}