package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})
	for _, currentJob := range jobs {
		wg.Add(1)
		go routinePipeline(in, out, wg, currentJob)
		//close(out)
		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}

func routinePipeline(in, out chan interface{}, wg *sync.WaitGroup, currJob job) {
	defer wg.Done()
	defer close(out)
	//defer close(in)
	currJob(in, out)
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for data := range in {
		wg.Add(1)
		go getSingleHash(data, out, wg, mu)
	}
	wg.Wait()
}

func getSingleHash(in interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	crc32Chan := make(chan string)
	dataInt, err := in.(int)
	if err {
		fmt.Println("in not integer")
	}
	dataStr := strconv.Itoa(dataInt)

	go getCrc32Async(dataStr, crc32Chan)

	//в задании мы не можем из 2 потоков обратиться к DataSignerMd5 иначе улетаем по времени
	mu.Lock()
	md5val := DataSignerMd5(dataStr)
	mu.Unlock()

	//go getCrc32Async(md5val, crc32Chan)
	//md5Data := <-crc32Chan
	md5Data := DataSignerCrc32(md5val)
	crcData := <-crc32Chan
	close(crc32Chan)
	out <- crcData + "~" + md5Data
}

func getCrc32Async(in string, out chan string) {
	out <- DataSignerCrc32(in)
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		go getMultiHash(data, out, wg)
	}
	wg.Wait()

}

func getMultiHash(in interface{}, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	additionalWG := &sync.WaitGroup{}
	//crc32chan := make(chan string, 6)
	dataStr, err := in.(string)
	if err {
		fmt.Println("in not string")
	}
	// через канал значения кладутся рандомно, чего нам не надо
	var resultArr [6]string

	result := ""
	for i := 0; i < 6; i++ {
		val := strconv.Itoa(i) + dataStr
		//go getCrc32Async(val, crc32chan) никаких каналов
		additionalWG.Add(1)
		go setValue(&resultArr[i], val, additionalWG)

	}
	additionalWG.Wait()
	result = strings.Join(resultArr[:], ``)
	out <- result
}

func setValue(val *string, strVal string, wg *sync.WaitGroup) {
	defer wg.Done()
	*val = DataSignerCrc32(strVal)
}

func CombineResults(in, out chan interface{}) {
	//fmt.Println("combine")
	result := make([]string, 0, 6)
	for value := range in {
		strVal, err := value.(string)
		if err {
			fmt.Println("value not string")
		}

		result = append(result, strVal)
	}
	sort.Strings(result)
	if len(result) == 0 {
		out <- ""
		return
	}

	resultString := strings.Join(result, `_`)

	out <- resultString
}
