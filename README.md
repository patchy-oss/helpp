# helpp
helpp - a CLI tool that helps you work without leaving your terminal

## Usage
```bash
helpp how do write a file from cli
helpp "How do I convert webm to mp4 using ffmpeg?"
helpp golang how do i open a file
```

## Dependencies
This program uses Google Gemini as a backend for user queries and requires an API key.
You can obtain the key [here](https://aistudio.google.com/app/apikey).

## How to install (linux & mac)
First, you need to build the program:
```bash
go mod tidy
go build
```

After that, you need to install it and set the appropriate ENV variables in the confg file:
```bash
go install
# helpp sets up the initial configs for you on the first run
helpp
# set the API key and (optionally) change the model
$EDITOR $HOME/.config/helpp/env
```
