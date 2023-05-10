package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "strconv"
)

func main() {
    // Check if the required arguments are provided
    if len(os.Args) < 5 {
        log.Fatal("Usage: chop [inputFile] [startTime] [length] [outputFile]")
    }

    // Input parameters
    inputFile := os.Args[1]
    startTime, err := strconv.Atoi(os.Args[2])
    if err != nil {
        log.Fatal("Invalid start time. Please provide an integer value.")
    }
    length, err := strconv.Atoi(os.Args[3])
    if err != nil {
        log.Fatal("Invalid length. Please provide an integer value.")
    }
    outputFile := os.Args[4]

    // Check if the input file exists
    if _, err := os.Stat(inputFile); os.IsNotExist(err) {
        log.Fatalf("Input file '%s' does not exist", inputFile)
    }

    // Determine the number of streams in the input file
    streamCount, err := getStreamCount(inputFile)
    if err != nil {
        log.Fatalf("Failed to determine the number of streams: %v", err)
    }

    // Build FFmpeg command
    command := []string{"-i", inputFile}
    filterComplex := fmt.Sprintf("[0:v]trim=0:%d,setpts=PTS-STARTPTS[v1];[0:v]trim=%d,setpts=PTS-STARTPTS[v2]", startTime, startTime+length)
    outputMap := "[v1]"
    if streamCount > 1 {
        filterComplex += fmt.Sprintf(";[0:a]atrim=0:%d,asetpts=PTS-STARTPTS[a1];[0:a]atrim=%d,asetpts=PTS-STARTPTS[a2]", startTime, startTime+length)
        outputMap += "[a1]"
    }
    filterComplex += fmt.Sprintf(";[v1][v2]concat=n=2:v=1%s", outputMap)
    command = append(command, "-filter_complex", filterComplex, "-map", outputMap, outputFile)

    // Execute FFmpeg command
    ffmpegCmd := exec.Command("ffmpeg", command...)
    ffmpegCmd.Stderr = os.Stderr
    ffmpegCmd.Stdout = os.Stdout
    if err := ffmpegCmd.Run(); err != nil {
        log.Fatalf("Failed to run FFmpeg command: %v", err)
    }

    fmt.Println("Video trimming completed successfully!")
}

// getStreamCount determines the number of streams (video and audio) in the input file.
func getStreamCount(inputFile string) (int, error) {
    command := exec.Command("ffprobe", "-i", inputFile, "-show_entries", "stream=codec_type", "-of", "compact=p=0:nk=1")
    output, err := command.Output()
    if err != nil {
        return 0, fmt.Errorf("failed to execute ffprobe: %v", err)
    }

    streamCount := 0
    for _, stream := range output {
        if stream == 'v' || stream == 'a' {
            streamCount++
        }
    }

    return streamCount, nil
}

