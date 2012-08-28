

package main

import (
    "ebooker"
    "bufio"
    "fmt"
    "io"
    "os"
    "flag"
)

func main() {
    filename := flag.String("file", "laurelita-sample.txt",
        "file containing corpus texts, seperated by line")
    prefixLen := flag.Int("prefixLength", 1, "length of prefix")

    flag.Parse()

    file, err := os.Open(*filename)
    if err != nil { panic(err) }
    defer file.Close()

    reader := bufio.NewReader(file)

    gen := ebooker.CreateGenerator(*prefixLen, 140)
    for {
        str, err := reader.ReadString(0x0A) // 0x0A == '\n'
        if err == io.EOF {
            break
        } else if err != nil {
            panic(err)
        }

        gen.AddSeeds(str)

    }

    fmt.Println(gen.GenerateText())

    fetcher := ebooker.CreateTweetFetcher("laurelita")
    fetcher.GetUserTimeline()
}
