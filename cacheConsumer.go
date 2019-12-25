package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
)

func main() {
    ch := make(chan string)
    for i := 0; i < 10; i++ {
        go func() {
            resp, _ := http.Get("http://localhost:4445/cache")
            defer resp.Body.Close()

            body, _ := ioutil.ReadAll(resp.Body)
            fmt.Println(string(body))
        }()
    }
    <-ch
}

