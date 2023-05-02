package configserver

import (
	"fmt"
	"github.com/kkdai/youtube/v2"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func DownloadYouTubeVideo(videoURL string, outputDir string) (string, error) {
	client := youtube.Client{}

	var format youtube.Format
	videoID, _ := youtube.ExtractVideoID(videoURL)
	video, err := client.GetVideo(videoID)
	if err != nil {
		return  "", fmt.Errorf("get video: %v", err)
	}

	format = *getSmallestFileSizeFormat(video.Formats)
	fmt.Printf("Downloading Format %s\n", format)
	stream, _, err := client.GetStream(video, &format)
	if err != nil {
		return "", fmt.Errorf("get stream: %v", err)
	}

	outputFilePath := filepath.Join(outputDir, video.Title+".mp4")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("create output file: %v", err)
	}

	_, err = io.Copy(outputFile, stream)
	if err != nil {
		return "", fmt.Errorf("save video: %v", err)
	}

	fmt.Printf("Video saved to %s\n", outputFilePath)
	return outputFilePath, nil
}


func getSmallestFileSizeFormat(formats []youtube.Format) *youtube.Format {
	if len(formats) == 0 {
		return nil
	}

	sortedFormats := make([]youtube.Format, len(formats))
	copy(sortedFormats, formats)

	sort.Slice(sortedFormats, func(i, j int) bool {
		return sortedFormats[i].ContentLength < sortedFormats[j].ContentLength
	})

	return &sortedFormats[0]
}
