package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type Status struct {
	Wind  int `json:"wind"`
	Water int `json:"water"`
}

type Weather struct {
	Status Status `json:"status"`
}

type MyFile struct {
	Name           string
	ModifiedNotify chan bool
}

// Modify updates file content and notify the channel
func (mf MyFile) Modify(content Weather) {
	jsonData, _ := json.Marshal(content)

	err := os.WriteFile(mf.Name, jsonData, fs.ModePerm)
	if err != nil {
		log.Fatalln("error while writing file :", err.Error())
	}
	go func() {
		mf.ModifiedNotify <- true
	}()
}

// Read reads and return the file content
func (mf MyFile) Read() []byte {
	content, err := os.ReadFile(mf.Name)
	if err != nil {
		log.Fatalln("error while reading file :", err.Error())
	}

	return content
}

func (mf MyFile) StartLoopUpdate(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)

	for {
		select {
		case <-ticker.C:
			mf.Modify(Weather{
				Status: Status{
					Wind:  GenerateRandomNumber(1, 100),
					Water: GenerateRandomNumber(1, 100),
				},
			})
		}
	}
}

func GenerateRandomNumber(min, max int) int {
	return rand.Intn(max-min) + min
}

var (
	FILENAME string
	ADDRESS  = ":3000"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	fName := flag.String("filename", "weather", "name of the file to store the weather data")

	flag.Parse()

	FILENAME = *fName + ".json"
}

func main() {
	// file info to keep track the updates
	myFile := &MyFile{
		Name:           FILENAME,
		ModifiedNotify: make(chan bool),
	}

	// start updating weather file constantly for given interval
	go myFile.StartLoopUpdate(time.Second * 15)

	// page route
	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		content := myFile.Read()

		var data map[string]any

		json.Unmarshal(content, &data)

		templ := template.Must(template.ParseFiles("pages/weather.html"))

		templ.Execute(w, data)
	})

	fmt.Println(fmt.Sprintf("Server running on %s", ADDRESS))
	log.Fatalln(http.ListenAndServe(ADDRESS, nil))
}
