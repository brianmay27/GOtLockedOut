package main

import (
   "github.com/kidoman/embd"
   _"github.com/kidoman/embd/host/rpi"
)

func initI2C(chans channels) {
   b := embd.NewI2CBus(1)
   l := NewLcd(b)
   go l.run(chans)
   k := NewKeypad(b)
   go k.run(chans)
   l.defaultScreen()
}
