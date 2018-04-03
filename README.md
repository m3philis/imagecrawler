#imagecrawler for konachan.com
================

## Why this crawler?

This little project was created to easily download pictures of Konachan.
The WebUi of Konachan let you see single pictures, but if you want to download pictures based on tags, you need to right-click every single one of them to save them to disk.
This tool just wants to know the tags, the path to save and if safemode should be used. After that it starts to use the konachan API to download all pictures it finds.


## Usage

Just use `go run` to run the tool without compiling.
`go run imagecrawler.go --dir foo --tags bar,snafu --safe`


### Flags

The crawler has 3 flags to be used:

`--dir`: Defines the directory where to save pictures. Default is `%HOME/pictures/konachan/unnamed`. The `--dir` option defines the `unnamed` at the end. The `%HOME/pictures/konachan` is hardcoded at the moment.
`--tags`: A coma seperated list of tags to search for (--tags foo,bar,snafu,height:1080)
`--safe`: A boolean to enable safemode (Default is off)
