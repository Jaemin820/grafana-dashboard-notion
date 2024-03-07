package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Dashboard struct {
	ID          int      `json:"id"`
	UID         string   `json:"uid"`
	Title       string   `json:"title"`
	Tags        []string `json:"tags"`
	URL         string   `json:"url"`
	Type        string   `json:"type"`
	FolderTitle string   `json:"folderTitle"`
}

func main() {
	// log 형식 세팅
	log.SetFlags(log.Ltime | log.LstdFlags | log.Llongfile)

	// csv 파일 초기 생성
	outputFile, err := os.Create("[Detect] Dashboard 장부.csv")
	if err != nil {
		log.Println(err)
		return
	}
	defer outputFile.Close()

	// csv 파일에 내용을 쓸 writer 생성
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// .env 파일 로드
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Grafana API Key 불러오기
	ALERT_RULES_READ_ONLY_API_KEY := os.Getenv("ALERT_RULES_READ_ONLY_API_KEY")

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://nodeinfra.grafana.net/api/search?query=&", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+ALERT_RULES_READ_ONLY_API_KEY)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error on request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Printf("Error: received status code %d\n", resp.StatusCode)
		return
	}

	var searchResults []Dashboard
	err = json.Unmarshal(body, &searchResults)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	dashboards := filterDashboards(searchResults)

	// Write CSV header with FolderTitle
	writer.Write([]string{"Folder", "Title", "Tags", "UID", "URL"})

	for _, d := range dashboards {
		tags := strings.Join(d.Tags, ",")
		writer.Write([]string{d.FolderTitle, d.Title, tags, d.UID, fmt.Sprintf("https://nodeinfra.grafana.net%s", d.URL)})
	}

	fmt.Println("Filtered dashboard information exported to filtered_dashboards.csv successfully.")
}

// filterDashboards filters the dashboards to include only those with type "dash-db"
func filterDashboards(dashboards []Dashboard) []Dashboard {
	var filtered []Dashboard
	for _, d := range dashboards {
		if d.Type == "dash-db" {
			filtered = append(filtered, d)
		}
	}
	return filtered
}
