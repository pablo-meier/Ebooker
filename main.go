

package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
    "markov"
)

func main() {
    filename := "laurelita-sample.txt"

    file, err := os.Open(filename)
    if err != nil { panic(err) }
    defer file.Close()

    reader := bufio.NewReader(file)

    gen := markov.CreateGenerator(1, 140)
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
}
