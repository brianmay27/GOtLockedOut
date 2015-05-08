package main

import (
	"fmt"
   "strings"
   "runtime"
   "os/signal"
   "os"
   "log"
   "syscall"
   "strconv"
   "bufio"
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
      signal.Notify(sigs, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL)
      buf := make([]byte, 1<<20)
      for {
         s := <-sigs
         if s == syscall.SIGQUIT {
            runtime.Stack(buf, true)
            log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf)
         } else {
            fmt.Print("**********EXITING**********")
            os.Exit(1)
         }
      }
   }()
	go initF(chans)
   go initI2C(chans)
   d := initD()
   defer d.Close()
   users := initU()
   go func() {
      var inp string
      for {
         fmt.Print(">")
         reader := bufio.NewReader(os.Stdin)
         text, err := reader.ReadString('\n')
         if err !=  nil {
            continue
         }
         s := strings.Split(text, " ")
         fmt.Println(inp)
         switch {
            case s[0] == "enroll":
               fmt.Println("enrolling")
               if len(s) != 4 {
                  continue
               }
               id, _ := strconv.Atoi(s[1])
               name := s[2]
               pin := []byte(s[3])
               usr := user{name, id, pin}
               users.add(usr)
               fmt.Println(usr)
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
   var id int
   for {
      outLoop:
      select {
         case finger := <-chans.fOut:
            fmt.Println("correct finger print")
            chans.kIn <- 0x00
            id = finger.id
            break;
         case user := <-chans.kOut:
            usr := users.get(id)
            if usr.fid == 0 {
               blog(2, "Id not found in users")
               break;
            }
            for i, n := range *user.pin {
               if n != usr.pin[i] {
                  break outLoop
               }
            }
            d.unlock()
            fmt.Println("Welcome in: ", usr.name)
            break;
      }
   }
}
