package main

import (
   "fmt"
   "github.com/kidoman/embd"

)
const (
   addressLcd = 0x48
)

type lcdInfo struct {
   pin int
   status byte
}

type lcd struct {
   Bus embd.I2CBus
   pin int
}

func NewLcd(b embd.I2CBus) *lcd {
   return &lcd{b,0}
}

func (d *lcd) run(chans channels) {
   for {
      c := <-chans.lIn
      if c.status == 1{
         d.clear()
         d.writeBytes([]byte("Place finger"))
      } else if c.status == 2 {
         d.clear()
         d.writeBytes([]byte("Replace finger"))
      } else if c.status == 3 {
         d.clear()
         d.writeBytes([]byte("Success"))
      } else if c.pin == 0 {
         d.pinDefault()
      } else if c.pin >= 7 {
        d.defaultScreen() 
      } else {
         d.writePin(c.pin)
      }
   }
}

func (d *lcd) pinDefault() {
   d.clear()
   d.writeCmd([]byte{'0',';','3','H'})
   d.writeBytes([]byte("Enter pin"))
   d.pin = 0
}

func (d *lcd) writePin(l int) {
   loc := 2+(2*(l-1))
   if d.pin > l {
      loc += 2
   }
   if loc > 9 {
      loc -= 10
      d.writeCmd([]byte{'1', ';','1', byte(loc + 0x30), 'H'})
   } else {
      d.writeCmd([]byte{'1', ';', byte(loc + 0x30), 'H'})
   }
   if d.pin > l {
      d.write(' ')
      d.pin--
   } else {
      d.write('*')
      d.pin++
   }
}

func (d *lcd) defaultScreen() {
   d.clear()
   d.writeCmd([]byte{'0', ';', '4', 'H'})
   d.writeBytes([]byte("Welcome"))
}

func (d *lcd) clear() {
   d.writeBytes([]byte{0x1B, '[', 'j'})
}

func (d *lcd) writeCmd(c []byte) {
   d.writeBytes(append([]byte{0x1B, '['}, c...))
}

func (d *lcd) writeBytes(c []byte) {
   err := d.Bus.WriteBytes(addressLcd, c)
   if err != nil {
      fmt.Println("Can not communicate with Lcd")
   }
}

func (d *lcd) write(c byte) {
   err := d.Bus.WriteByte(addressLcd, c)
   if err != nil {
      fmt.Println("Can not communicate with Lcd")
   }
}
