package main

import (
    "log"
    "fmt"
    "net/http"
    "io/ioutil"
)

func main() {
    ch := make(chan string)
    for i := 0; i < 10; i++ {
        go func() {
            resp, err := http.Get("http://localhost:4445/cache")
            if err != nil { log.Fatal(err) }
            defer resp.Body.Close()

            body, _ := ioutil.ReadAll(resp.Body)
            fmt.Println(string(body))
        }()
    }
    <-ch
}

