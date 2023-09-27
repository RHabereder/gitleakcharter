/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/spf13/cobra"
)

var inputFileDir string
var outputFile string
var allFiles []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "leakcharter",
	Short: "Creates fancy bar charts for Gitleaks Reports",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		walkAllFiles(inputFileDir, ".json")
		generateBarChart(inputFileDir, outputFile)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	rootCmd.Flags().StringVarP(&inputFileDir, "file", "f", "report.json", "Input report to feed into the Chart-Generator")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "chart.html", "Output-Filename")
	//rootCmd.MarkFlagFilename("file", ".json")
	rootCmd.MarkFlagDirname("file")
	rootCmd.MarkFlagRequired("file")
	rootCmd.MarkFlagRequired("output")
}

func walkAllFiles(directory string, extension string) {
	filepath.Walk(inputFileDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == extension {
			allFiles = append(allFiles, path)
		}
		return nil
	})
}

func removeDuplicateValues(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
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

		file, err := os.Open(file)
		defer file.Close()

		stat, err := file.Stat()
		bs := make([]byte, stat.Size())
		_, err = bufio.NewReader(file).Read(bs)
		if err != nil && err != io.EOF {
			fmt.Println(err)
		}

		if err := json.Unmarshal([]byte(bs), &reportItems); err != nil {
			panic(err)
		}

		for _, item := range reportItems {
			allKeys = append(allKeys, item.RuleID)
		}
	}
	return removeDuplicateValues(allKeys)
}

func generateBarChart(inputFile string, outputFile string) {
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

		file, err := os.Open(file)
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

		for _, item := range reportItems {
			output[item.RuleID] = append(output[item.RuleID], item)
		}

		sort.Strings(removeDuplicateValues(allKeys))
		barDatas[file.Name()] = append(generateBarLeakItems(allKeys, output))

	}

	for key, data := range barDatas {
		bar.AddSeries(key, data, charts.WithLabelOpts(opts.Label{Show: true}))
	}

	modifiedKeys := removeDuplicateValues(allKeys)
	bar.SetXAxis(modifiedKeys)

	// Where the magic happens
	f, _ := os.Create(outputFile)
	bar.Render(f)
}

// generate random data for bar chart
func generateBarLeakItems(keys []string, reportItems map[string][]ReportItem) []opts.BarData {
	items := make([]opts.BarData, 0)
	for _, k := range keys {
		items = append(items, opts.BarData{Value: len(reportItems[k]), Name: k})
	}
	return items
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
