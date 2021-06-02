package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"runtime"
	"sync"
	"time"
)


func main(){
	start := time.Now()

	WordCounter(ioutil.Discard)

	elapsed := time.Since(start)
	log.Printf("WordCounter took %s", elapsed)

}

//function for searching the slice of bytes in the slice of slice of bytes
func isUsed(arr *[][]byte, word *[]byte) int{
	for i := 0; i < len(*arr); i++ {
		if bytes.Compare((*arr)[i], *word) == 0 {
			return i
		}
	}
	return -1
}

func WordCounter(out io.Writer) {

	data, err := ioutil.ReadFile(".\\mobydick.txt")

	if err != nil {
		fmt.Println(err)
	}

	//one dimensional array for storing a single word
	var oned []byte
	//kind of sorted slice. In few words, it stores words
	var sortedSlice [][]byte

	size := len(data)

	for i := 0; i < size-1; i++ {
		//checking here whether a byte is a letter or a symbol
		if data[i] >= 97 && data[i] <= 122 || data[i] >= 65 && data[i] <= 90 {
			//and appending only symbols
			oned = append(oned, data[i])
			//if array does not find any letters it means that new word started
			continue
		}
		if len(oned) > 0 {
			//empty array check
			sortedSlice = append(sortedSlice, oned)
		}
		oned = []byte{}
	}

	size = len(sortedSlice)

	fmt.Println(size)

	numOfCPUs := runtime.NumCPU();
	//Slice for checked words, reading and counting already checked words cause huge overhead

	var wg sync.WaitGroup
	var mutex sync.Mutex

	type matcher struct {
		byteArray []byte
		occurence uint
	}

	var matcherChannel = make(chan matcher, size)

	//var wordChannel = make(chan []byte, size)
	//var occurrenceChannel = make(chan uint, size)

	if numOfCPUs > 1 {

		avgNumberOfLines := int(math.Round(float64(size / numOfCPUs)))

		for i:=0; i<numOfCPUs;i++ {

			wg.Add(1);

			if diff:=size - avgNumberOfLines;diff < 0 {
				avgNumberOfLines-=diff;
			}

			size-=avgNumberOfLines

			go func(matchChannel chan matcher,
				wg *sync.WaitGroup, mutex *sync.Mutex,
				text [][]byte, numberOfLines int, firstIndex int, lastIndex int) {

				defer wg.Done()

				var usedWords [][]byte
				var occurrenceSlice []uint

				usedWords = append(usedWords, text[firstIndex])
				occurrenceSlice = append(occurrenceSlice, 1)

				for i := firstIndex+1; i<=lastIndex; i++{
					index := isUsed(&usedWords, &text[i])
					if index == -1{
						usedWords = append(usedWords, text[i])
						occurrenceSlice = append(occurrenceSlice, 1)
					}else {
						occurrenceSlice[index] += 1
					}
				}

				fmt.Println("UW - ", len(usedWords))

				for i:=0; i<len(usedWords); i++ {
					matcherChannel <- matcher{
						byteArray: usedWords[i],
						occurence: occurrenceSlice[i],
					}
				}

			}(matcherChannel, &wg, &mutex,
				sortedSlice, avgNumberOfLines,
				i*avgNumberOfLines, (i+1)*avgNumberOfLines-1)

		}

	}

	var matchedArray []matcher;
	counter := 0

	go func(matChannel chan matcher, matchedArr *[]matcher, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		for request := range matChannel{
			counter++
			if length:=len(matchedArray); length > 0 && matchedArray != nil {

				for i:= 0; i < length; i++ {
					if bytes.Compare((*matchedArr)[i].byteArray, request.byteArray) == 0 {
						(*matchedArr)[i].occurence+=request.occurence
						return
					}
				}

				matchedArray = append(matchedArray, matcher{
					byteArray: request.byteArray,
					occurence: request.occurence})

			} else {

				matchedArray = append(matchedArray, matcher{
					byteArray: request.byteArray,
					occurence: request.occurence})

			}
		}

	}(matcherChannel, &matchedArray, &wg)

	//go func(ocChannel chan uint) {
	//	for request := range ocChannel{
	//		fmt.Println("occurrence ", request)
	//	}
	//}(occurrenceChannel)
	wg.Wait()

	for i:= 0; i < len(matchedArray); i++ {
		fmt.Println(string(matchedArray[i].byteArray), ":", matchedArray[i].occurence)
	}

	fmt.Println(len(matchedArray))
	//usedWords = append(usedWords, sortedSlice[0])
	//occurrenceSlice = append(occurrenceSlice, 1)
	//
	//for i := 1; i< size; i++{
	//	index = isUsed(&usedWords, &sortedSlice[i])
	//	if index == -1{
	//		usedWords = append(usedWords, sortedSlice[i])
	//		occurrenceSlice = append(occurrenceSlice, 1)
	//	}else {
	//		occurrenceSlice[index] += 1
	//	}
	//}
	//
	//size = len(occurrenceSlice)
	//
	////bubble sort for sorting arrays by occurrence
	//for i := 0; i < size-1; i++ {
	//	for j := i+1; j < size-i-1; j++ {
	//		if occurrenceSlice[j] > occurrenceSlice[j+1] {
	//
	//			temp := occurrenceSlice[j]
	//			occurrenceSlice[j] = occurrenceSlice[j+1]
	//			occurrenceSlice[j+1] = temp
	//
	//			byteSlice := usedWords[j]
	//			usedWords[j] = usedWords[j+1]
	//			usedWords[j+1] = byteSlice
	//
	//		}
	//	}
	//}
	//
	////printing used words
	//for i := 0; i < 25; i++ {
	//	print(string(usedWords[size-i-1]) + " ")
	//	println(occurrenceSlice[size-i-1])
	//}
	//fmt.Fprintln(out, "fuck")
}