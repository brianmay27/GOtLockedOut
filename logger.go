package main

import (
   "fmt"
)

func blog(level int, a ...interface{}) {
   if DEBUG {
      if level <= DEBUG_LEVEL {
         fmt.Println(a)
      }
   }
}

