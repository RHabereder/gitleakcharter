package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/wcharczuk/go-chart/v2"
)

// generate random data for bar chart
func generateBarLeakItems(keys []string, reportItems map[string][]ReportItem) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, k := range keys {
		items = append(items, opts.BarData{Value: len(reportItems[k]), Name: k})
	}
	return items
}

func createPieChart(keys []string, reportItems map[string][]ReportItem) {
	chartValues := make([]chart.Value, 0)

	for _, k := range keys {
		chartValues = append(chartValues, chart.Value{Value: float64(len(reportItems[k])), Label: k})
	}

	pie := chart.PieChart{
		Width:  512,
		Height: 512,
		Values: chartValues,
	}

	f, _ := os.Create(os.Getenv("LC_PIE_OUTPUT_FILE"))
	defer f.Close()
	pie.Render(chart.PNG, f)
}

func main() {
	var reportItems []ReportItem
	file, err := os.Open(os.Getenv("LC_INPUT_FILE"))
	defer file.Close()

	stat, err := file.Stat()
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		fmt.Println(err)
		return
	}

	if err := json.Unmarshal([]byte(bs), &reportItems); err != nil {
		panic(err)
	}

	output := make(map[string][]ReportItem)
	for _, item := range reportItems {
		output[item.RuleID] = append(output[item.RuleID], item)
	}

	// Sort Output Map
	keys := make([]string, 0, len(output))

	for k := range output {
		keys = append(keys, k)
	}
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: "Gitleak-Charter",
	}))

	bar.Initialization.Width = "1200"

	// Put data into instance
	sort.Strings(keys)
	bar.SetXAxis(keys).
		AddSeries("Amount of Leaktype", generateBarLeakItems(keys, output))

	// Where the magic happens
	f, _ := os.Create(os.Getenv("LC_BAR_OUTPUT_FILE"))
	bar.Render(f)

	createPieChart(keys, output)

}

type ReportItem struct {
	Description string    `json:"Description"`
	StartLine   int       `json:"StartLine"`
	EndLine     int       `json:"EndLine"`
	StartColumn int       `json:"StartColumn"`
	EndColumn   int       `json:"EndColumn"`
	Match       string    `json:"Match"`
	Secret      string    `json:"Secret"`
	File        string    `json:"File"`
	SymlinkFile string    `json:"SymlinkFile"`
	Commit      string    `json:"Commit"`
	Entropy     float64   `json:"Entropy"`
	Author      string    `json:"Author"`
	Email       string    `json:"Email"`
	Date        time.Time `json:"Date"`
	Message     string    `json:"Message"`
	Tags        []any     `json:"Tags"`
	RuleID      string    `json:"RuleID"`
	Fingerprint string    `json:"Fingerprint"`
}
