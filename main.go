package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	dotenv "github.com/joho/godotenv"
)

type Entry struct {
	PeakSpeed string `json:"peak_speed"`
	AvgSpeed  string `json:"avg_speed"`
	TimeTaken int64  `json:"time_taken"`
}

type Benchmark struct {
	Version  string `json:"version"`
	Layer    int    `json:"layer"`
	Download Entry  `json:"download"`
	Upload   Entry  `json:"upload"`
}

func main() {
	var benchmark = Benchmark{
		Version: tg.Version,
		Layer:   tg.ApiVersion,
	}

	dotenv.Load()
	var (
		APP_ID       = os.Getenv("TG_API_ID")
		APP_HASH     = os.Getenv("TG_API_HASH")
		BOT_TOKEN    = os.Getenv("TG_BOT_TOKEN")
		MESSAGE_LINK = os.Getenv("TG_MESSAGE_LINK")
	)

	appIdInt, _ := strconv.Atoi(APP_ID)

	client, _ := tg.NewClient(tg.ClientConfig{
		AppID:         int32(appIdInt),
		AppHash:       APP_HASH,
		LogLevel:      tg.LogInfo,
		MemorySession: true,
		DisableCache:  true,
	})
	client.LoginBot(BOT_TOKEN)
	messageId, _ := strconv.Atoi(strings.Split(MESSAGE_LINK, "/")[4])
	message, _ := client.GetMessageByID(strings.Split(MESSAGE_LINK, "/")[3], int32(messageId))

	var prog = tg.NewProgressManager(2)
	var peakSpeed int64 = 0
	var startTime int64 = time.Now().Unix()
	var fileSize int64 = message.File.Size

	prog.Edit(func(totalSize, currentSize int64) {
		if time.Now().Unix()-startTime == 0 {
			return
		}
		var currSpeed = currentSize / (time.Now().Unix() - startTime)
		if currSpeed > peakSpeed {
			peakSpeed = currSpeed
		}
	})

	downloaded, _ := message.Download(&tg.DownloadOptions{
		ProgressManager: prog,
	})

	var avgSpeed = float64(fileSize) / float64(time.Now().Unix()-startTime)

	benchmark.Download = Entry{
		PeakSpeed: HumanizeBytes(peakSpeed),
		AvgSpeed:  HumanizeBytes(int64(avgSpeed)),
		TimeTaken: time.Now().Unix() - startTime,
	}

	prog = tg.NewProgressManager(2)
	peakSpeed = 0
	startTime = time.Now().Unix()
	fileSize = message.File.Size

	prog.Edit(func(totalSize, currentSize int64) {
		if time.Now().Unix()-startTime == 0 {
			return
		}
		var currSpeed = currentSize / (time.Now().Unix() - startTime)
		if currSpeed > peakSpeed {
			peakSpeed = currSpeed
		}
	})

	client.UploadFile(downloaded, &tg.UploadOptions{
		ProgressManager: prog,
	})

	avgSpeed = float64(fileSize) / float64(time.Now().Unix()-startTime)

	benchmark.Upload = Entry{
		PeakSpeed: HumanizeBytes(peakSpeed),
		AvgSpeed:  HumanizeBytes(int64(avgSpeed)),
		TimeTaken: time.Now().Unix() - startTime,
	}

	jsonBenchmark, _ := json.MarshalIndent(benchmark, "", "  ")
	os.WriteFile("benchmark.json", jsonBenchmark, 0644)
}

func HumanizeBytes(size int64) string {
	var units = []string{"B", "KB", "MB", "GB", "TB"}
	var i = 0
	for size > 1024 {
		size = size / 1024
		i++
	}
	return fmt.Sprintf("%.2f %s/s", float64(size), units[i])
}
