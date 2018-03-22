package main

// import "fmt"
// import "time"

func getMessagesChannel() <-chan int {
     c := make(chan int)

     go func() {
       day := 7
       index6 := 6
       var lst [7]int
       for i:= 0; i < 7; i++ {
         lst[i] = 1
       }

       // fmt.Printf("%v", lst)
       // println("\n")

       total := 0

       for {
         num := lst[day-1] + lst[day-7] * 8
         // f(x) = f(day-1) + f(day-7) * 8
         // fmt.Printf("total: %v", total)
         // println("\n")

         // shift lst left by one and set lst[6] = n
         for j := 0; j < 6; j++ {
           lst[j] = lst[j+1]
         }
         lst[index6] = num
         total += num

         // fmt.Printf("%v", lst)
         // println("\n")

         c <- lst[index6]
       }
     }()
     return c
}

func main() {
     c1 := getMessagesChannel()
     for i:= 0; i < 10; i++ {
          println(<-c1)
     }
}
