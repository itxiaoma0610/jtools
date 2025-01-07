package main

import (
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
)

func CheckLinkValid(jobs <-chan []string, validCsvWriter *csv.Writer, inValidCsvWriter *csv.Writer, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 忽略 SSL 证书验证（仅用于测试环境）
			},
			DisableCompression: true, // 禁用压缩
			MaxConnsPerHost:    200,  // 增加对每个主机的最大连接数
		},
	}

	for job := range jobs {
		url := strings.TrimSpace(job[4])
		// 创建请求并添加 User-Agent 头
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Error creating request for URL %s: %v", url, err)
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:133.0) Gecko/20100101 Firefox/133.0")

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 400 {
			// log.Printf("Error accessing URL %s: %v", url, err)
			mu.Lock()
			if err := inValidCsvWriter.Write(job); err != nil {
				log.Printf("Error writing to invalid CSV: %v", err)
			}
			mu.Unlock()
			continue
		} else {
			mu.Lock()
			if err := validCsvWriter.Write(job); err != nil {
				log.Printf("Error writing to valid CSV: %v", err)
			}
			mu.Unlock()
		}

		if resp != nil {
			resp.Body.Close()
		}
	}
}

func Execute(filePath string) {
	startTime := time.Now()

	const numJobs = 120
	connChannel := make(chan []string, numJobs)

	var wg sync.WaitGroup
	var mu sync.Mutex

	records, validWriter, inValidWriter, err := CheckCSVFileExists(filePath, "good.csv", "bad.csv")
	if err != nil {
		log.Fatalf("CheckCSVFileExists error: %v", err)
		return
	}
	defer validWriter.Flush()
	defer inValidWriter.Flush()

	// 启动100协程
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go CheckLinkValid(connChannel, validWriter, inValidWriter, &wg, &mu)
	}
	bar := progressbar.NewOptions(len(records),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[JTools] Process file..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	for _, row := range records {
		connChannel <- row
		bar.Add(1)
	}

	close(connChannel)
	wg.Wait()

	fmt.Printf("花费时间：%v\n", time.Since(startTime))
}

func main() {
	ticker := time.NewTicker(7 * 24 * time.Hour) // 七天自动执行一次
	defer ticker.Stop()                          // 程序退出时停止定时器

	var filePath string
	flag.StringVar(&filePath, "f", "Unknown", "请输入文件路径")

	flag.Parse()

	Execute(filePath)

	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			Execute(filePath)
		case <-exitSignal:
			fmt.Println("程序退出")
			return
		}
	}
}
