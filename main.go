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
		for (*words)[left].Occurrence > pivot {
			left++
		}
		for (*words)[right].Occurrence < pivot {
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
	//reading bytes from file
	chars, err := ioutil.ReadFile(".\\mobydick.txt")
	if err != nil {
		fmt.Println(err)
	}
	var word []byte
	var words [][]byte
	size := len(chars)
	//get all words from chars and store in words array
	//run until size - 1, because we don't want to include dot -> .
	for i := 0; i < size - 1; i++ {
		//if char is a letter
		if chars[i] >= 65 && chars[i] <= 90{
			word = append(word, chars[i] + 32)
		} else if chars[i] >= 97 && chars[i] <= 122 {
			word = append(word, chars[i])
		} else {
			//if char is not a letter -> word ends
			if len(word) > 0 {
				words = append(words, word)
				word = []byte{}
			}
		}
	}
	//we calculate number of CPUs in order to run goroutines
	numOfCPUs := runtime.NumCPU()
	var wg sync.WaitGroup
	var matcherChannel = make(chan []Matcher)
	size = len(words)
	/*
		This is our part size. Example:
		8 CPUs. Words count - 74817. Then, we need to run each goroutine for equal part, divide 74817/8 - avgNumberOfLines
	 */
	avgNumberOfLines := int(math.Floor(float64(size)/float64(numOfCPUs)))
	wg.Add(numOfCPUs)
		for i := 0; i < numOfCPUs; i++ {
			//while size is more than avgNumberofLines -> decrease size
			if size - avgNumberOfLines > 0 {
				size -= avgNumberOfLines
			}
			//for first word of part
			firstIndex := i * avgNumberOfLines
			//for last word of part
			secondIndex := (i + 1) * avgNumberOfLines - 1
			//if it's last goroutine, then we work with all remaining words -> secondIndex will be equal to initial size
			if i + 1 == numOfCPUs {
				secondIndex += size
			}
			go func(matchChannel chan []Matcher, wg *sync.WaitGroup,
				text [][]byte, firstIndex int, lastIndex int) {

				defer wg.Done()
				//our words cache
				var matchArray []Matcher
				//we don't need to check matchArray for nil inside loop, so we append first word
				matchArray = append(matchArray, Matcher{
					Word:       text[firstIndex],
					Occurrence: 1,
				})
				//run loop, but from the second word inside the part
				for i := firstIndex + 1; i <= lastIndex; i++ {
					//trying to find word in our cache
					index := indexOfWord(&matchArray, &text[i])
					//if not present
					if index == -1 {
						//add word to cache
						matchArray = append(matchArray, Matcher{
							Word:       text[i],
							Occurrence: 1,
						})
					} else {
						//increment occurrence of word in cache
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
		//our words, but with counted occurrence
	var finalWords []Matcher
	for request := range matcherChannel {
		finalWords = append(finalWords, request...)
	}
	//our cache, but for finalWords
	var cache []Matcher
	var index int
	//as in previous go func, we append first word with its occurence
	cache = append(cache, Matcher{
		Word:       finalWords[0].Word,
		Occurrence: finalWords[0].Occurrence,
	})
	size = len(finalWords)
	//our task is to remove all duplicate words and store distinct words with their occurrence inside second cache
	for i := 1; i < size; i++{
			//find word, which matches any words in second cache
			index = indexOfMatcher(&cache, &finalWords[i])
			if index == -1 {
				cache = append(cache, finalWords[i])
			} else {
				(cache)[index].Occurrence += finalWords[i].Occurrence
			}
	}
	size = len(cache)
	//we have got all needed distinct words with occurence, so sort them
	sort(&cache, 0, len(cache) - 1)
	//finally, outputting first 25 words with biggest occurrences
	for i := 0; i < 25; i++ {
		print(string(cache[i].Word) + " ")
		println(cache[i].Occurrence)
	}
}