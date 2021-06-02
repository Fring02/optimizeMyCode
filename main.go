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

func main() {
	start := time.Now()
	WordCounter(ioutil.Discard)
	elapsed := time.Since(start)
	log.Printf("WordCounter took %s", elapsed)

}

type Matcher struct {
	Word []byte
	Occurrence uint
}
func sort(words *[]Matcher, first int, last int) {
	left, right := first, last
	pivot := (*words)[(left+right)/2].Occurrence
	for left <= right {
		for (*words)[left].Occurrence < pivot {
			left++
		}
		for (*words)[right].Occurrence > pivot {
			right--
		}
		if left <= right {
			(*words)[left], (*words)[right] = (*words)[right], (*words)[left]
			left++
			right--
		}
	}
	if first < right {
		sort(words, first, right)
	}
	if left < last {
		sort(words, left, last)
	}
}
func indexOfWord(arr *[]Matcher, word *[]byte) int {
	for i := 0; i < len(*arr); i++ {
		if bytes.Equal((*arr)[i].Word, *word) {
			return i
		}
	}
	return -1
}
func indexOfMatcher(arr *[]Matcher, word *Matcher) int {
	for i := 0; i < len(*arr); i++ {
		if bytes.Equal((*arr)[i].Word, (*word).Word) {
			return i
		}
	}
	return -1
}
func WordCounter(out io.Writer) {
	chars, err := ioutil.ReadFile(".\\mobydick.txt")
	if err != nil {
		fmt.Println(err)
	}
	var word []byte
	var words [][]byte
	size := len(chars)
	//get all words from chars and store in words array
	for i := 0; i < size - 1; i++ {
		//if char is a letter
		if chars[i] >= 97 && chars[i] <= 122 || chars[i] >= 65 && chars[i] <= 90 {
			word = append(word, chars[i])
		} else {
			//if char is not a letter -> word ends
			if len(word) > 0 {
				words = append(words, word)
				word = []byte{}
			}
		}
	}
	numOfCPUs := runtime.NumCPU()
	var wg sync.WaitGroup
	var matcherChannel = make(chan []Matcher)
	size = len(words)
	avgNumberOfLines := int(math.Floor(float64(size)/float64(numOfCPUs)))
	wg.Add(numOfCPUs)
		for i := 0; i < numOfCPUs; i++ {
			if size - avgNumberOfLines > 0 {
				size -= avgNumberOfLines
			}
			firstIndex := i * avgNumberOfLines
			secondIndex := (i + 1) * avgNumberOfLines - 1
			if i + 1 == numOfCPUs {
				secondIndex += size
			}
			go func(matchChannel chan []Matcher, wg *sync.WaitGroup,
				text [][]byte, firstIndex int, lastIndex int) {
				defer wg.Done()
				var matchArray []Matcher
				matchArray = append(matchArray, Matcher{
					Word:       text[firstIndex],
					Occurrence: 1,
				})
				for i := firstIndex + 1; i <= lastIndex; i++ {
					index := indexOfWord(&matchArray, &text[i])
					if index == -1 {
						matchArray = append(matchArray, Matcher{
							Word:       text[i],
							Occurrence: 1,
						})
					} else {
						matchArray[index].Occurrence++
					}
				}
				matcherChannel <- matchArray
			}(matcherChannel, &wg, words, firstIndex, secondIndex)
		}
	go func() {
		wg.Wait()
		close(matcherChannel)
	}()
	var finalWords []Matcher
	for request := range matcherChannel {
		finalWords = append(finalWords, request...)
	}
	var cache []Matcher
	var index int
	cache = append(cache, Matcher{
		Word:       finalWords[0].Word,
		Occurrence: finalWords[0].Occurrence,
	})
	size = len(finalWords)
	for i := 1; i < size; i++{
			index = indexOfMatcher(&cache, &finalWords[i])
			if index == -1 {
				cache = append(cache, finalWords[i])
			} else {
				(cache)[index].Occurrence += finalWords[i].Occurrence
			}
	}
	size = len(cache)
	sort(&cache, 0, len(cache) - 1)
	for i := 0; i < 25; i++ {
		print(string(cache[size-i-1].Word) + " ")
		println(cache[size-i-1].Occurrence)
	}
}