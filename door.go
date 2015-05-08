package main

import (
   "github.com/kidoman/embd"
   "time"
   "fmt"
)

type door struct {
   pin embd.DigitalPin
}

func initD() *door {
   embd.InitGPIO()
   pin, err := embd.NewDigitalPin(17)
   if err != nil {
      fmt.Println(err)
   }
   d := door{pin}
   d.pin.SetDirection(embd.Out)
   return &d
}

func (d *door) Close() {
   embd.CloseGPIO()
}

func (d *door) unlock() {
   err := d.pin.Write(embd.High)
   if err != nil {
      fmt.Println(err)
   }
   fmt.Println("Unlocking")
   go func(d *door) {
      c := time.After(8*time.Second)
      <-c
      d.lock()
   }(d)
}

func (d *door) lock() {
   d.pin.Write(embd.Low)
   fmt.Println("Locked")
}
