package main

import (
   _"fmt"
   "os"
   "log"
   "strings"
   "bufio"
   "strconv"
   "bytes"
   "fmt"
   "io/ioutil"
)

type user struct {
   name string
   fid int
   pin []byte
}

type allUsers struct {
   users map[int]user
}

func initU() *allUsers {
   o, err := os.Open("users.txt")
   defer o.Close()
   if err != nil {
      log.Fatal(err)
   }
   u := allUsers{make(map[int]user)}
   sca := bufio.NewScanner(o)
   for sca.Scan() {
      text := sca.Text()
      s := strings.Split(text, ";")
      if len(s) != 3 {
         break
      }
      fmt.Println(s)
      fid, _ := strconv.Atoi(s[1])
      pin := make([]byte, 6)
      for i, n := range []byte(s[2]) {
         pin[i] = n - 0x30
      }
      u.users[fid] = user{s[0], fid, pin}
   }
   return &u
}

func (u *allUsers) add(us user) {
   u.users[us.fid] = us
   writeToFile(u)
}

func (u *allUsers) get(id int) user {
   return u.users[id]
}

func (u *allUsers) update(us user) {

}

func writeToFile(usr *allUsers) {
   buf := bytes.NewBufferString("")
   if DEBUG {
      for _, val := range usr.users {
         fmt.Fprintf(buf, "%s;%d;%s\n", val.name, val.fid, val.pin)
      }
   }
   err := ioutil.WriteFile("users.txt", buf.Bytes(), 0640)
   if err == nil {
      blog(2, "Unable to write user file")
      return
   }
}
