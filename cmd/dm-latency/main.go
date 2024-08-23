package main

import "sync"

func main() {
	monitor := Monitor{}
	if err := monitor.Init(); err != nil {
		panic(err)
	}
	generator := Generator{}
	if err := generator.Init(); err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	//go monitor.Start(wg)
	go generator.Start(wg)
	wg.Wait()
}
