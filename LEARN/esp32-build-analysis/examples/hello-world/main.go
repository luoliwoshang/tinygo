package main

import (
	"machine"
	"time"
)

func main() {
	// 配置内置LED
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	println("Hello from TinyGo on ESP32!")
	println("LED blinking started...")

	// 无限循环闪烁LED
	for {
		led.Low()  // LED点亮 (ESP32上LED低电平有效)
		time.Sleep(500 * time.Millisecond)
		
		led.High() // LED熄灭
		time.Sleep(500 * time.Millisecond)
		
		println("LED toggle")
	}
}