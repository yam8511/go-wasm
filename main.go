package main

import (
	"context"
	"fmt"
	"log"
	"syscall/js"
	"time"
)

func getElementByID(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}

var mainCtx, ctx context.Context
var mainExit, stop context.CancelFunc

func startSpin(this js.Value, args []js.Value) interface{} {
	if ctx != nil {
		return false
	}

	go func() {
		log.Println("開始每秒spin")
		ctx, stop = context.WithCancel(mainCtx)
		count := 0
		for {
			count++
			select {
			case <-time.After(time.Second):
				v := getElementByID("score").Get("value").String()
				log.Printf("第%d局 => 換分比: %s", count, v)
			case <-ctx.Done():
				return
			}
		}
	}()
	return true
}

func stopSpin(this js.Value, args []js.Value) interface{} {
	if stop == nil {
		return false
	}

	stop()
	log.Println("結束spin")
	ctx = nil
	stop = nil
	return true
}

func sayHello(this js.Value, args []js.Value) interface{} {
	fmt.Println("Hello Wasm!")
	return true
}

func setGlobalMethods() {
	js.Global().Set("say_hello", js.FuncOf(sayHello))
	js.Global().Set("start_spin", js.FuncOf(startSpin))
	js.Global().Set("stop_spin", js.FuncOf(stopSpin))
}

func init() {
	fmt.Println("Init Global")
	setGlobalMethods()
}

func main() {
	mainCtx, mainExit = context.WithCancel(context.Background())

	exitButton := getElementByID("exitButton")
	startButton := getElementByID("startButton")
	stopButton := getElementByID("stopButton")

	//// Situation 2
	// runButton := getElementByID("runButton")
	// runButton.Set("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
	// 	js.Global().Call("run")
	// 	return nil
	// }))
	// runButton.Set("disabled", true)

	exitButton.Set("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		startButton.Set("disabled", true)
		startButton.Set("hidden", true)
		stopButton.Set("disabled", true)
		stopButton.Set("hidden", true)
		exitButton.Set("disabled", true)
		//// Situation 2
		// runButton.Set("disabled", false)
		mainExit()
		return nil
	}))
	exitButton.Set("disabled", false)

	startButton.Set("onclick", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		startButton.Set("disabled", true)
		stopButton.Set("disabled", false)
		stopButton.Set("hidden", false)
		js.Global().Call("start_spin")
		return nil
	}))
	startButton.Set("disabled", false)
	startButton.Set("hidden", false)

	stopButton.Set("onclick", js.FuncOf(func(js.Value, []js.Value) interface{} {
		js.Global().Call("stop_spin")
		stopButton.Set("hidden", true)
		stopButton.Set("disabled", true)
		startButton.Set("disabled", false)
		return nil
	}))

	fmt.Println("App is ready go.")
	<-mainCtx.Done()
	fmt.Println("App exit.")

	time.Sleep(time.Second)
}
