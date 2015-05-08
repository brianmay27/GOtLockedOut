package main

import (
   "github.com/tarm/serial"
   "time"
   "log"
   "fmt"
   "io"
   "errors"
   "image"
   "image/color"
   "image/png"
   "os"
)


type fingerInfo struct{
   status int
   id int
}
type fingerP struct {
   io io.ReadWriteCloser
   ch *channels
}
const addr = 0xffffffff
const minConf = 100

func initF(chans channels) {
   c := & serial.Config{Name: "/dev/ttyAMA0", Baud: 115200, ReadTimeout: time.Second * 5}
   fmt.Println("opening")
   enabled := true
   s, err := serial.OpenPort(c)
   if err != nil {
      log.Fatal("Cant open serial port")
   }
   p := fingerP{s, &chans}
   go func() {
      for {
         todo := <-chans.fIn
         fmt.Println(todo)
         switch {
            case todo.status == 0 :
               enabled = false
               break
            case todo.status == 1 :
               enabled = true
               break
            case todo.status == 10:
               enabled = false
               status := p.enroll(0x01)
               fmt.Println(status)
               enabled = true
               break
            case todo.status == 100:
               enabled = false
               p.deleteUsers()
               enabled=true
               break
            case todo.status == 50:
               p.getImage()
               b := p.getPicture()
               t := time.Now()
               filet := fmt.Sprintf("%d-%2d-%2d:%2d:%2d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
               img, _ := os.Create("/home/pi/log/" + filet + ".png")
               defer img.Close()
               png.Encode(img, b )
               break
         }
      }
   }()

   //Now that Im using a ReadTimeout, I should not need to worry about it becoming hung
/*   t := time.NewTimer(time.Duration(10)*time.Second)

   go func(ch <-chan time.Time) {
      for {
         <-ch
         fmt.Println("watchdog called")
         s.Close()
         s, err = serial.OpenPort(c)
         if err != nil {
            fmt.Println("Error opening serial")
         }
      }
   }(t.C)*/
   for {
      if !enabled {
         time.Sleep(time.Duration(500)*time.Millisecond)
         //t.Reset(time.Duration(10)*time.Second)
         continue
      }
      img := p.getImage()
      if img == FINGERPRINT_OK {
         img = p.image2Tz(1)
         if img == FINGERPRINT_OK {
            id, con := p.search()
            if con > minConf {
               fmt.Println("User id: ", id)
               chans.fOut <- fingerInfo{0, id}
               chans.lIn <- lcdInfo{0, 0}
            }
            //t.Reset(time.Duration(15)*time.Second)
            img := p.getPicture()
            if img == nil {
               chans.lIn <- lcdInfo{0,5}
               continue
            }
            t := time.Now()
            filet := fmt.Sprintf("%d-%2d-%2d:%2d:%2d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
            file, _ := os.Create("/home/pi/log/" + filet + ".png")
            defer file.Close()
            png.Encode(file, img )

         }
      }
      //t.Reset(time.Duration(10)*time.Second)
      time.Sleep(time.Duration(2)*time.Millisecond)
   }

}

func (p *fingerP) writePacket(packetType byte, packet []byte) error {
   _, err := p.io.Write([]byte{byte(FINGERPRINT_STARTCODE>>8), byte(FINGERPRINT_STARTCODE&0xff) , 0xff, 0xff, 0xff, 0xff , packetType , byte((len(packet)>>8)), byte(len(packet)&0xff + 2)})
   if err != nil {
      fmt.Println("Error writing to serial")
      return err
   }
   sum := byte(len(packet)>>8) + byte(len(packet)&0xff) + packetType
   for i := range packet {
      sum += packet[i]
   }
   _, err = p.io.Write(append(packet, []byte{sum>>8, sum&0xff + 2}...))
   if err != nil {
      fmt.Println("Error writing to serial")
      return err
   }
   return nil
}

func (p *fingerP) readPacket(rec []byte) ([]byte, error) {
   b := make([]byte, 9)
   var n int
   var err error
   var c []byte
   for i :=0 ; true ; i++ {
      n, err = p.io.Read(b)
      if err !=  nil {
         blog(2, "Cant read from UART:", err)
         return nil, err
      }
      b = b[:n]
      if c != nil {
         b = append(c, b...)
      }
      if len(b) >= 9 {
         break
      } else {
         c = b[:n]
         b = make([]byte, 9 - n)
      }
      if i > 50 {
         blog(3, "Packet capture failed, unable to read enough packets")
         return nil, errors.New("Unable to read enough bits")
      }
   }
   if b[0] != FINGERPRINT_STARTCODE>>8 || b[1] != FINGERPRINT_STARTCODE&0xff {
      blog(4, "startcode is invalid", b)
      return nil, errors.New("Invalid start code")
   }
   length := int(b[7]<<8 | b[8])
   status := b[6]
   blog(6, "Packet status: ", status)
   rest := make([]byte, length)
   time.Sleep(time.Duration(((length*4000)/14400))*time.Millisecond)
   n, err = p.io.Read(rest)
   if n != length {
      blog(3, "Unable to retreve all data", b, rest, n)
   }
   b = append(b, rest...)
   b = b[9:len(b)-2]
   if status == 0x02 {
      if rec != nil {
         b = append(rec, b...)
      }
      blog(6, "Continue to get data")
      b, err = p.readPacket(b)
      if err != nil {
         blog(2, "Multiple packet capture failed")
      }
      return b, nil
   } else if status == 0x07 {
      //All good over here
   } else if status == 0x08{
      b = append(rec, b...)
      return b, nil
   } else {
      //Unknow reason it failed
      blog(4, "unknows status: ", status)
      return nil, errors.New("Unknown status: " + string(status))
   }
   blog(6, "packet: ", b)
   return b, nil
}

func (p *fingerP) getImage() byte {
   p.writePacket(FINGERPRINT_COMMANDPACKET, []byte{FINGERPRINT_GETIMAGE})
   c, err := p.readPacket(nil)
   if err != nil {
      return 0
   }
   return c[0]
}

func (p *fingerP) image2Tz(slot byte) byte {
   p.writePacket(FINGERPRINT_COMMANDPACKET, []byte{FINGERPRINT_IMAGE2TZ, slot})
   c, err := p.readPacket(nil)
   if err != nil {
      return 0
   }
   return c[0]
}

func (p *fingerP) search() (int, int) {
   p.writePacket(FINGERPRINT_COMMANDPACKET, []byte{FINGERPRINT_HISPEEDSEARCH, 0x01, 0x00, 0x00, 0x00, 0xA3})
   c, err := p.readPacket(nil)
   if err != nil {
      return 0,0
   }
   if len(c) < 5 {
      fmt.Print("recieved less than expected", c)
      return 0,0
   }
   if c[0] != 0x00 {
      blog(4, "Fingerprint not found")
      return 0, 0
   }
   fingerId := int((c[1]<<8) | c[2])
   confidence := int((c[3] <<8) | c[4])
   return fingerId, confidence
}

func (p *fingerP) createModel() byte {
   p.writePacket(FINGERPRINT_COMMANDPACKET, []byte{FINGERPRINT_REGMODEL})
   c, err := p.readPacket(nil)
   if err != nil {
      return 0
   }
   return c[0]
}

func (p *fingerP) storeModel(id int) byte {
   p.writePacket(FINGERPRINT_COMMANDPACKET, []byte{FINGERPRINT_STORE, byte(0x01), byte((id & 0xFF00) >> 8), byte(id & 0xFF)})
   c, err := p.readPacket(nil)
   if err != nil {
      return 0
   }
   return c[0]
}

func (p *fingerP) enroll(id int) error {
  t := time.NewTimer(time.Duration(15) * time.Second)
  for {
     go func(c <-chan time.Time) error {
        <-c
        return errors.New("Timed out")
     }(t.C)
     p.ch.lIn <- lcdInfo{0,1}
     p.getImageBlock(1)
     t.Reset(time.Duration(15) * time.Second)
     p.ch.lIn <- lcdInfo{0,2}
     time.Sleep(time.Duration(7)*time.Second)
     t.Reset(time.Duration(15) * time.Second)
     p.getImageBlock(2)
     t.Reset(time.Duration(15) * time.Second)
     i := p.createModel()
     if i != FINGERPRINT_OK {
        //Finger prints dont match
        t.Reset(time.Duration(15) * time.Second)
        continue
     }
     fmt.Println(id)
     i = p.storeModel(id)
     if i == FINGERPRINT_OK {
        p.ch.lIn <- lcdInfo{0,3}
        return nil
     }
  }
  return errors.New("Enrollment failed for unknown reason")
}

func (p *fingerP) getImageBlock(loc byte) {
   for {
      i := p.getImage()
      if i == FINGERPRINT_OK {
         i = p.image2Tz(loc)
         if i == FINGERPRINT_OK {
            return 
         }
      }
      time.Sleep(time.Duration(10)*time.Millisecond)
   }
}

func (p *fingerP) getPicture() image.Image {
   i := p.getImage()
   if i != FINGERPRINT_OK {
      return nil
   }
   p.writePacket(FINGERPRINT_COMMANDPACKET, []byte{0x0a})
   b, _ := p.readPacket(nil)
   b, _ = p.readPacket(nil)
   img := formatBmp(b)
   blog(4, "Successful image capture")
   return img
}

func formatBmp(arr []byte) image.Image {
   rect := image.Rect(0,0,256,288)
   img := image.NewGray(rect)
   for y:= 0; y < 288; y++ {
      for x:= 0; x < 128; x++ {
         pix := arr[128*y + x]
         g1 := pix & 0xF0;
         g2 := pix & 0x0F << 4
         gr := color.Gray{g1}
         gr2 := color.Gray{g2}
         img.SetGray(x*2,y,gr)
         img.SetGray(x*2 + 1, y, gr2)
      }
   }
   return img.SubImage(rect)
}


func (p *fingerP) deleteUsers() {
   p.writePacket(FINGERPRINT_COMMANDPACKET , []byte{0x0D})
   p.readPacket(nil)
}
