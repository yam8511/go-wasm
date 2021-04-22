package main

import (
	"context"
	"fmt"
	"log"
	"os"
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
			select {
			case <-ctx.Done():
				return
			default:
			}

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

func add(this js.Value, args []js.Value) interface{} {
	var sum float64
	for _, v := range args {
		if v.IsNaN() {
			continue
		}

		sum += v.Float()
	}
	return sum
}

func init() {
	fmt.Println("init")
	js.Global().Set("say_hello", js.FuncOf(sayHello))
	js.Global().Set("start_spin", js.FuncOf(startSpin))
	js.Global().Set("stop_spin", js.FuncOf(stopSpin))
	js.Global().Set("add_num", js.FuncOf(add))
	js.Global().Set("wasm", js.ValueOf(map[string]interface{}{
		"say_hello":  js.FuncOf(sayHello),
		"start_spin": js.FuncOf(startSpin),
		"stop_spin":  js.FuncOf(stopSpin),
		"add_num":    js.FuncOf(add),
	}))
}
func main() {
	mainCtx, mainExit = context.WithCancel(context.Background())
	fmt.Println("Args = ", os.Args)
	fmt.Println("Env = ", os.Environ())

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
}
