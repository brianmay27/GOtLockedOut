package main

import (
	"fmt"
   "strings"
   "runtime"
   "os/signal"
   "os"
   "log"
   "syscall"
)

type channels struct {
   fIn chan fingerInfo
   fOut chan fingerInfo
   kIn chan byte
   kOut chan keypadInfo
   lIn chan lcdInfo
}

func main() {
   chans := channels{fIn: make(chan fingerInfo),fOut: make(chan fingerInfo), kIn: make(chan byte), kOut: make(chan keypadInfo), lIn: make(chan lcdInfo)}
   go func() {
      sigs := make(chan os.Signal, 1)
      signal.Notify(sigs, syscall.SIGQUIT)
      buf := make([]byte, 1<<20)
      for {
         <-sigs
         runtime.Stack(buf, true)
         log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf)
      }
   }()
	go initF(chans)
   go initI2C(chans)
   go func() {
      var inp string
      for {
         fmt.Print(">")
         _, err := fmt.Scanln(&inp)
         if err !=  nil {
            continue
         }
         s := strings.Split(inp, " ")
         switch {
            case s[0] == "enroll":
               fmt.Println("enrolling")
               chans.fIn <- fingerInfo{10,0x01}
               break
            case s[0] == "disable":
               chans.fIn <- fingerInfo{0,0}
               break
            case s[0] == "enable":
               chans.fIn <- fingerInfo{1,0}
               break
            case s[0] == "delete!!!":
               chans.fIn <- fingerInfo{100, 0}
               break
            case s[0] == "capture":
               chans.fIn <- fingerInfo{50,0}
               break
         }
      }

   }()
   for {
      select {
         case <-chans.fOut:
            fmt.Println("correct finger print")
            chans.kIn <- 0x00
            
            break;
         case <-chans.kOut:
            
            break;
      }
   }

}
