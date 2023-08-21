package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "io"
  "io/ioutil"
  "log"
  "net/http"
  "net/url"
  "os"
  "os/user"
  "strconv"
  "strings"
  "sync"
  "time"
)

// Global vars
var wg sync.WaitGroup

type picture struct {
  FileURL string `json:"file_url"`
  Width        int    `json:"width"`
  Height       int    `json:"height"`
  ImageWidth   int    `json:"image_width"`
  ImageHeight  int    `json:"image_height"`
  FileExt      string `json:"file_ext"`
  Tags         string `json:"tag_string"`
}

// main function to download pictures
func main() {

  // define flags and parse them
  var path string
  var safemode bool
  var tags string
  var aspect string
  var site string

  // variables for downloading
  picHits := 1
  page := 1
  ratio := 0.0

  flag.StringVar(&path, "dir", "unnamed", "Directory to safe pictures. Default is %HOME/pictures/konachan/unnamed")
  flag.BoolVar(&safemode, "safe", false, "Safemode to filter NSFW pictures. Default is false")
  flag.StringVar(&tags, "tags", "", "Tags used to filter search query.")
  flag.StringVar(&aspect, "aspect", "", "Aspect ratio pics should have")
  flag.StringVar(&site, "site", "konachan", "Site to crawl from, either konachan or danbooru")

  flag.Parse()

  // set home directory and create it to save pictures in
  homepath, err := user.Current()
  if err != nil {
    log.Fatal(err)
  }
  filepath := strings.Join([]string{homepath.HomeDir, "pictures", site, strings.TrimSuffix(path, "\n")}, "/")
  os.MkdirAll(filepath, 0700)

  // edit tags array to met API requirement
  tags = strings.Replace(tags, ",", "+", -1)
  tags = strings.Replace(tags, "=", ":", -1)
  tags = strings.TrimSuffix(tags, "\n")

  // calculate aspect ratio
  if isFlagPassed("aspect") {
    aspectSlice := strings.Split(aspect, ":")
    widthF, _ := strconv.ParseFloat(aspectSlice[0], 64)
    heightF, _ := strconv.ParseFloat(aspectSlice[1], 64)
    ratio = widthF / heightF
  } else {
    ratio = 0.0
  }

  for picHits > 0 {

    fmt.Println("Page: ", page)

    website := fmt.Sprintf("https://konachan.com/post.json?page=%d&tags=%s", page, tags)
    if safemode {
      website = fmt.Sprintf("https://konachan.com/post.json?page=%d&tags=%s+rating:safe", page, tags)
    }

    if site == "danbooru" {
      website = fmt.Sprintf("https://danbooru.donmai.us/posts.json?page=%d&tags=%s", page, tags)
      if safemode {
        website = fmt.Sprintf("https://danbooru.donmai.us/posts.json?page=%d&tags=%s+rating:safe", page, tags)
      }
    }

    picList := openConnection(website)
    pictures, count := parseMaps(picList, ratio)

    picHits = count
    page++

    wg.Add(len(pictures))
    for _, pic := range pictures {
      go downloadPic(pic, filepath)
    }
    wg.Wait()
    time.Sleep(1 * time.Second)
  }
}

func isFlagPassed(name string) bool {
  found := false
  flag.Visit(func(f *flag.Flag) {
    if f.Name == name {
      found = true
    }
  })
  return found
}

// function to create the connection to konachan and get the API response
func openConnection(url string) []picture {
  var f []picture

  result, err := http.Get(url)
  if err != nil {
    log.Fatal(err)
  }
  defer result.Body.Close()

  data, err := ioutil.ReadAll(result.Body)
  if err != nil {
    log.Fatal(err)
  }

  if err = json.Unmarshal(data, &f); err != nil {
    panic(err)
  }

  return f
}

// function to parse the json response and extract only the file url
func parseMaps(f []picture, ratio float64) ([]string, int) {
  fileURLs := []string{}
  picCount := 0

  if isFlagPassed("aspect") {
    for _, pic := range f {
      picCount++

      picWidthF := 1.0
      picHeightF := 1.0

      if (pic.Width != 0 && pic.Height != 0) {
        picWidthF = float64(pic.Width)
        picHeightF = float64(pic.Height)
      } else {
        picWidthF = float64(pic.ImageWidth)
        picHeightF = float64(pic.ImageHeight)
      }

      if (picWidthF / picHeightF) == ratio {
        fileURL := pic.FileURL
        fileURLs = append(fileURLs, fileURL)
      }
    }
  } else {
    for _, pic := range f {
      picCount++
      fileURL := pic.FileURL
      fileURLs = append(fileURLs, fileURL)
    }
  }
  return fileURLs, picCount
}

// function to download and sace the pictures to disk
func downloadPic(picURL string, filepath string) {
  defer wg.Done()

  picName, err := url.PathUnescape(strings.Split(picURL, "/")[len(strings.Split(picURL, "/"))-1])
  if err != nil {
    log.Fatal(err)
  }

  if _, err := os.Stat(filepath + "/" + picName); err == nil {
    return
  }

  result, err := http.Get(picURL)
  if err != nil {
    log.Fatal(err)
  }
  defer result.Body.Close()

  //fmt.Println(result.Status)

  if result.StatusCode != 200 {
    wg.Add(1)
    go downloadPic(picURL, filepath)
    return
  }

  file, err := os.Create(filepath + "/" + picName)
  if err != nil {
    log.Fatal(err)
  }

  _, err = io.Copy(file, result.Body)
  if err != nil {
    log.Fatal(err)
  }

  file.Close()

  fmt.Printf("Downloading: %s\n", picName)
}
