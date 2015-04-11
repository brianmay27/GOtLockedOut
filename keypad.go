package main

import (
   "fmt"
   "github.com/kidoman/embd"
   "errors"
   "time"
)

const(
   addressKeypad = 0x70
)

type keypadInfo struct {
   pin *[]byte
   values int
}

type keypad struct {
   Bus embd.I2CBus
   enabled bool
}

func NewKeypad(d embd.I2CBus) *keypad {
   return &keypad{d, false}
}

func (d *keypad) defaultScreen() {

}

func (d *keypad) enable() error {
   err := d.write(0x01)
   if err != nil {
      return err
   }
   d.enabled = true
   return nil
}

func (d *keypad) disable() {
   d.write(0x05)
   d.enabled = false
}

func (d *keypad) run(chans channels) {
   for {
      <-chans.kIn
      if !d.enabled {
         fmt.Println("Getting pin")
         err := d.enable()
         if err != nil {
            continue
         }
         b := make([]byte, 6)
         err = d.write(0x04)
         for i := 0; i < 6; {
            t, err := d.Bus.ReadByte(addressKeypad) 
            if err != nil {
               fmt.Println("error reading keypad")
               continue
            }
            if t == 0x23 || t == 0x2A {
               i--
               chans.lIn<-lcdInfo{i,0}
               chans.kOut<-keypadInfo{&b, i}
            } else if t != byte(0xff) {
               fmt.Printf("Got pin: %x\n", t)
               b[i] = t
               i++
               chans.lIn<-lcdInfo{i,0}
               chans.kOut<-keypadInfo{&b, i}
            }
            time.Sleep(time.Duration(100)*time.Millisecond)
         }
         d.disable()
         chans.lIn<-lcdInfo{7,0}
         chans.kOut<-keypadInfo{&b, 6}
      }
   }
}

func (d *keypad) write(c byte) error {
   tries := 0
   for err := errors.New("err"); err != nil; { 
      err = d.Bus.WriteByte(addressKeypad, c)
      if err != nil {
         tries++
         fmt.Println("Can not communicate with Keypad")
      }
      if tries > 5 {
         return err
      }
   }
   return nil
}

