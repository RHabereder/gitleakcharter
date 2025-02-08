// Package cmd Implements cobra commands for the CLI
package cmd

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/spf13/cobra"
)

var inputFileDir string
var outputFile string
var fileExtension string
var allFiles []string
var Logger log.Logger

var rootCmd = &cobra.Command{
	Use:    "leakcharter",
	PreRun: toggleDebug,
	Short:  "Creates fancy bar charts for Gitleaks Reports",
	Run: func(cmd *cobra.Command, args []string) {
		walkAllFiles(inputFileDir, "."+fileExtension)
		generateBarChart(outputFile)
	},
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&inputFileDir, "file", "f", "./reports/", "Directory with reports to feed into the Chart-Generator")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "chart.html", "Output-Filename")
	rootCmd.Flags().StringVarP(&fileExtension, "extension", "e", "json", "Extension of Report files to scan for")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "verbose logging")
	err := rootCmd.MarkFlagDirname("file")
	if err != nil {
		log.Fatal(err)
	}
	err = rootCmd.MarkFlagRequired("file")
	if err != nil {
		log.Fatal(err)
	}
	err = rootCmd.MarkFlagRequired("output")
	if err != nil {
		log.Fatal(err)
	}
}

func walkAllFiles(directory string, extension string) {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == extension {
			allFiles = append(allFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func removeDuplicateValues(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func readAllKeys(allFiles []string) []string {
	var allKeys []string
	for _, file := range allFiles {
		var reportItems []ReportItem

		file, _ := os.Open(file)
		defer file.Close()

		stat, _ := file.Stat()
		bs := make([]byte, stat.Size())
		_, err := bufio.NewReader(file).Read(bs)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}

		if err := json.Unmarshal([]byte(bs), &reportItems); err != nil {
			log.Panic(err)
		}

		for _, item := range reportItems {
			allKeys = append(allKeys, item.RuleID)
		}
	}
	return removeDuplicateValues(allKeys)
}

func generateBarChart(outputFile string) {
	barDatas := make(map[string][]opts.BarData, len(allFiles))

	bar := charts.NewBar()
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title: "Gitleak-Charter",
	}))
	bar.SetGlobalOptions(charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Interval: "0", Show: true, ShowMinLabel: true, ShowMaxLabel: true}}))
	bar.Initialization.Width = "500"

	allKeys := readAllKeys(allFiles)

	for _, file := range allFiles {
		output := make(map[string][]ReportItem)

		var reportItems []ReportItem
		log.Debug("Reading file " + file)
		file, _ := os.Open(file)
		defer file.Close()

		stat, _ := file.Stat()
		bs := make([]byte, stat.Size())
		_, err := bufio.NewReader(file).Read(bs)
		if err != nil && err != io.EOF {
			log.Fatal(err)
			return
		}

		log.Debug("Unmarshalling " + file.Name())
		if err := json.Unmarshal([]byte(bs), &reportItems); err != nil {
			log.Panic(err)
		}

		log.Debug("Unmarshalled " + strconv.Itoa(len(reportItems)) + " ReportItems")
		for _, item := range reportItems {
			output[item.RuleID] = append(output[item.RuleID], item)
		}

		sort.Strings(removeDuplicateValues(allKeys))
		barDatas[file.Name()] = generateBarLeakItems(allKeys, output)

	}

	for key, data := range barDatas {
		bar.AddSeries(key, data, charts.WithLabelOpts(opts.Label{Show: true}))
	}

	modifiedKeys := removeDuplicateValues(allKeys)
	bar.SetXAxis(modifiedKeys)

	f, _ := os.Create(outputFile)
	err := bar.Render(f)
	if err != nil {
		log.Fatal(err)
	}
}

func generateBarLeakItems(keys []string, reportItems map[string][]ReportItem) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, k := range keys {
		items = append(items, opts.BarData{Value: len(reportItems[k]), Name: k})
	}
	return items
}

// ReportItem is a Struct for unmarshalling Gitleaks Report Items
// for easier handling than raw JSON parsing
type ReportItem struct {
	RuleID      string        `json:"RuleID"`
	Description string        `json:"Description"`
	StartLine   int           `json:"StartLine"`
	EndLine     int           `json:"EndLine"`
	StartColumn int           `json:"StartColumn"`
	EndColumn   int           `json:"EndColumn"`
	Match       string        `json:"Match"`
	Secret      string        `json:"Secret"`
	File        string        `json:"File"`
	SymlinkFile string        `json:"SymlinkFile"`
	Commit      string        `json:"Commit"`
	Entropy     float32       `json:"Entropy"`
	Author      string        `json:"Author"`
	Email       string        `json:"Email"`
	Date        string        `json:"Date"`
	Message     string        `json:"Message"`
	Tags        []interface{} `json:"Tags"`
	Fingerprint string        `json:"Fingerprint"`
}
