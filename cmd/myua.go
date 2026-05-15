package cmd

import (
	"fmt"
	"io"
	"net/http"
)


func myua() {
    // make an HTTP GET request
    response, err := http.Get("https://httpbin.io/user-agent")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer response.Body.Close()

    // read the response body
    body, err := io.ReadAll(response.Body)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    // print the text content
    fmt.Println(string(body))
}
