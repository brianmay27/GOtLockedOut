package main

import (
   _"fmt"
   "os"
   "log"
   "strings"
   "bufio"
   "strconv"
)

type user struct {
   name string
   fid int
   pin string
}

type allUsers struct {
   users map[int]user
}

func getUsers() *allUsers {
   o, err := os.Open("users.txt")
   if err != nil {
      log.Fatal(err)
   }
   u := allUsers{make(map[int]user)}
   sca := bufio.NewScanner(o)
   for sca.Scan() {
      text := sca.Text()
      s := strings.Split(text, ";")
      fid, _ := strconv.Atoi(s[1])
      u.users[fid] = user{s[0], fid, s[2]}
   }
   o.Close()
   return &u
}

func (u *allUsers) add(us user) {

}

func (u *allUsers) update(us user) {

}
