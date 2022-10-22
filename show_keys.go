package main

import (
	"fmt"
	"sync"
)

func showKeys(pvtKeys []string, wgrp *sync.WaitGroup) {
	for idx, key := range pvtKeys {
		if idx+1 == len(pvtKeys) {
			fmt.Printf("%s\n", key)
		} else {
			fmt.Printf("%s, ", key)
		}
	}
	fmt.Println()
	wgrp.Done()
}
