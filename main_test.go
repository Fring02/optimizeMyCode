package main

import (
	"io/ioutil"
	"testing"
)

func BenchmarkWordCounter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WordCounter(ioutil.Discard)
	}
}