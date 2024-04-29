# glollama

ollama + pleasantly rendered markdown at the CLI.

# dependencies

* [Ollama](https://github.com/ollama/ollama)
* Go (to build/run this, no binaries are offered)

# run it

`go run main.go`

# choosing a model

glollama defaults to using the `codellama` model. specify any model that ollama support by setting the `MODEL` env var: `MODEL=llama3 go run main.go`
