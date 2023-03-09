package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
	"io"
	"net/url"
	"strconv"

	"main/jsbeautifier"
	
	"github.com/alexflint/go-arg"
	"github.com/dustin/go-humanize"
	"github.com/PuerkitoBio/goquery"
)

const (
	userAgent 			= "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKi"+
		"t/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"
	tracksArrayRegexStr = `tracks=(\[(?:{"track":\d+,"name":"(?:[^"]+|)","le`+
		`ngth":"\d+:\d+","file":"[^"]+"},)+])`
	urlRegexStr 		= `^https://downloads.khinsider.com/game-soundtracks/al`+
		`bum/([a-z0-9]+(?:-[a-z0-9]+)*)$`
	trackUrlBase 		=  "https://vgmsite.com/soundtracks/"
	sanRegexStr 		= `[\/:*?"><|]`
)

var (
	client 				= &http.Client{Transport: &Transport{}}
	tracksArrayReplacer = strings.NewReplacer(
		`\'\'`, `\"`, `\'`,  "'", "&#8203;", "")
)

var supportedFmts = []string{
	"FLAC", "M4A", "OGG", "MP3",
}

var resolveFmt = map[int]string{
	1: "MP3",
	2: "OGG",
	3: "M4A",
	4: "FLAC",
}

var fmtFallback = map[string]string{
	"FLAC":  "M4A",
	"M4A": "FLAC",
	"OGG": "MP3",
	"MP3": "OGG",
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(
		"User-Agent", userAgent,
	)
	// req.Header.Add(
	// 	"Referer", apiBase,
	// )
	return http.DefaultTransport.RoundTrip(req)
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	var speed int64 = 0
	n := len(p)
	wc.Downloaded += int64(n)
	percentage := float64(wc.Downloaded) / float64(wc.Total) * float64(100)
	wc.Percentage = int(percentage)
	toDivideBy := time.Now().UnixMilli() - wc.StartTime
	if toDivideBy != 0 {
		speed = int64(wc.Downloaded) / toDivideBy * 1000
	}
	fmt.Printf("\r%d%% @ %s/s, %s/%s ", wc.Percentage, humanize.Bytes(uint64(speed)),
		humanize.Bytes(uint64(wc.Downloaded)), wc.TotalStr)
	return n, nil
}

func handleErr(errText string, err error, _panic bool) {
	errString := errText + "\n" + err.Error()
	if _panic {
		panic(errString)
	}
	fmt.Println(errString)
}

func wasRunFromSrc() bool {
	buildPath := filepath.Join(os.TempDir(), "go-build")
	return strings.HasPrefix(os.Args[0], buildPath)
}

func getScriptDir() (string, error) {
	var (
		ok    bool
		err   error
		fname string
	)
	runFromSrc := wasRunFromSrc()
	if runFromSrc {
		_, fname, _, ok = runtime.Caller(0)
		if !ok {
			return "", errors.New("Failed to get script filename.")
		}
	} else {
		fname, err = os.Executable()
		if err != nil {
			return "", err
		}
	}
	return filepath.Dir(fname), nil
}

func readTxtFile(path string) ([]string, error) {
	var lines []string
	f, err := os.OpenFile(path, os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return lines, nil
}

func contains(lines []string, value string, fold bool) bool {
	for _, line := range lines {
		if fold {
			if strings.EqualFold(line, value) {
				return true
			}
		} else {
			if line == value {
				return true
			}
		}
	}
	return false
}

func processUrls(urls []string) ([]string, error) {
	var (
		processed []string
		txtPaths  []string
	)
	for _, _url := range urls {
		if strings.HasSuffix(_url, ".txt") && !contains(txtPaths, _url, true) {
			txtLines, err := readTxtFile(_url)
			if err != nil {
				return nil, err
			}
			for _, txtLine := range txtLines {
				if !contains(processed, txtLine, true) {
					txtLine = strings.TrimSuffix(txtLine, "/")
					processed = append(processed, txtLine)
				}
			}
			txtPaths = append(txtPaths, _url)
		} else {
			if !contains(processed, _url, true) {
				_url = strings.TrimSuffix(_url, "/")
				processed = append(processed, _url)
			}
		}
	}
	return processed, nil
}

func parseCfg() (*Config, error) {
	cfg, err := readConfig()
	if err != nil {
		return nil, err
	}
	args := parseArgs()
	if args.Format != -1 {
		cfg.Format = args.Format
	}
	if !(cfg.Format >= 1 && cfg.Format <= 4) {
		return nil, errors.New("Track format must be between 1 and 4.")
	}
	cfg.WantedFmt = resolveFmt[cfg.Format]
	if args.OutPath != "" {
		cfg.OutPath = args.OutPath
	}
	if cfg.OutPath == "" {
		cfg.OutPath = "KHInsider downloads"
	}
	if cfg.DiskNumPrefix == "" {
		cfg.DiskNumPrefix = "Disk "
	}
	cfg.Urls, err = processUrls(args.Urls)
	if err != nil {
		fmt.Println("Failed to process URLs.")
		return nil, err
	}
	return cfg, nil
}

func readConfig() (*Config, error) {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var obj Config
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func parseArgs() *Args {
	var args Args
	arg.MustParse(&args)
	return &args
}

func makeDirs(path string) error {
	err := os.MkdirAll(path, 0755)
	return err
}

func checkUrl(_url string) string {
	regex := regexp.MustCompile(urlRegexStr)
	match := regex.FindStringSubmatch(_url)
	if match == nil {
		return ""
	}
	return match[1]
}

func getDocument(_url string) (*goquery.Document, error) {
	resp, err := client.Get(_url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func filterFmts(fmts []string) []string {
	var filtered []string
	for _, fmt := range fmts {
		if !contains(supportedFmts, fmt, false) {
			filtered = append(filtered, fmt)
		}
	}
	return filtered
}


func getFname(file string) (string, error) {
	lastIdx := strings.LastIndex(file, "/")
	dec, err := url.QueryUnescape(file[lastIdx+1:])
	if err != nil {
		return "", err
	}
	lastIdx = strings.LastIndex(dec, ".")
	return sanitise(dec[:lastIdx+1]), nil
}

func extractMeta(_url, slug string) (*Meta, error) {
	doc, err := getDocument(_url)
	if err != nil {
		return nil, err
	}

	pageContent := doc.Find(`div[id="pageContent"]`).First()
	// Get this meta from the code instead of the table so wo don't have
	// to follow each row link for the other formats.
	options := jsbeautifier.DefaultOptions()
	code := pageContent.Find("script").First().Text()
	// Also unpacks so the code and meta's not garbled. The code doesn't get executed.
	unpackedCode, err := jsbeautifier.BeautifyString(code, options)
	if err != nil {
		return nil, err
	}
	regex := regexp.MustCompile(tracksArrayRegexStr)
	match := regex.FindStringSubmatch(*unpackedCode)
	tracksArrayStr := tracksArrayReplacer.Replace(match[1])
	tracksArrayStr = tracksArrayStr[:len(tracksArrayStr)-2] + "]"
	var tracks []*Track
	err = json.Unmarshal([]byte(tracksArrayStr), &tracks)
	if err != nil {
		return nil, err
	}

	var meta Meta
	meta.Tracks = tracks
	//fmt.Println(meta.Formats)
	var diskNum int
	for i, track := range meta.Tracks {
		if track.Track == 1 {
			diskNum ++
		}
		meta.Tracks[i].DiskNum = diskNum
		fname, err := getFname(track.File)
		if err != nil {
			return nil, err
		}
		if meta.Tracks[i].Name == "" {
			meta.Tracks[i].Name = "Track " + strconv.Itoa(i+1)
		}
		meta.Tracks[i].Fname = fname
		meta.Tracks[i].File = trackUrlBase+ slug +"/"+ track.File[:len(track.File)-3]
	}
	meta.HasDisks = diskNum > 1

	title := pageContent.Find("h2").First().Text()
	meta.Title = title
	tracklist := pageContent.Find(`table[id="songlist"]`).First()
	tracklistHeader := tracklist.Find(`tr[id="songlist_header"]`)

	var formats []string
	tracklistHeader.Find("th").Each(func(_ int, s *goquery.Selection) {
		// Row width reliable for formats?
		width, ok := s.Attr("width")
		if ok && width == "60px" {
			text := s.Text()
			if contains(supportedFmts, text, false) {
				formats = append(formats, text)
			} else {
				fmt.Println("Filtered unsupported format: " + text)
			}
			
		}
	})

	if len(formats) == 0 {
		return nil, errors.New("Formats array is empty.")
	}
	meta.Formats = formats

	return &meta, nil
}


func chooseFmt(formats []string, wantedFmt string) string {
	origWantedFmt := wantedFmt

	if len(formats) == 1 {
		wantedFmt = formats[0]
	} else {
		// MP3 always present?
		for {
			if contains(formats, wantedFmt, false) {
				break
			}
			if wantedFmt == "FLAC" && !contains(formats, wantedFmt, false) {
				if !contains(formats, "M4A", false) {
					wantedFmt = "OGG"
					continue
				}
			}
			if wantedFmt == "M4A" && !contains(formats, wantedFmt, false) {
				if !contains(formats, "FLAC", false) {
					wantedFmt = "OGG"
					continue
				}
			}			
			wantedFmt = fmtFallback[wantedFmt]
		}
	}

	if origWantedFmt != "FLAC" && wantedFmt != origWantedFmt {
		fmt.Println("Unavailable in you chosen format.")
	}
	return wantedFmt
}

func sanitise(filename string) string {
	regex := regexp.MustCompile(sanRegexStr)
	return regex.ReplaceAllString(filename, "_")
}

func fileExists(path string) (bool, error) {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir(), nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func downloadTrack(trackPath, _url string) error {
	f, err := os.OpenFile(trackPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := client.Get(_url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return errors.New(resp.Status)
	}

	totalBytes := resp.ContentLength
	counter := &WriteCounter{
		Total:     totalBytes,
		TotalStr:  humanize.Bytes(uint64(totalBytes)),
		StartTime: time.Now().UnixMilli(),
	}
	_, err = io.Copy(f, io.TeeReader(resp.Body, counter))
	
	fmt.Println("")
	return err
}

func init() {
	fmt.Println(`
 _____ _____ _____         _   _            ____                _           _         
|  |  |  |  |     |___ ___|_|_| |___ ___   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___ 
|    -|     |-   -|   |_ -| | . | -_|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|__|__|__|__|_____|_|_|___|_|___|___|_|    |____/|___|_____|_|_|_|___|__,|___|___|_|    
`)
}

func main() {
	scriptDir, err := getScriptDir()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(scriptDir)
	if err != nil {
		panic(err)
	}
	cfg, err := parseCfg()
	if err != nil {
		handleErr("Failed to parse config/args.", err, true)
	}
	err = makeDirs(cfg.OutPath)
	if err != nil {
		handleErr("Failed to make output folder.", err, true)
	}

	albumTotal := len(cfg.Urls)
	for albumNum, _url := range cfg.Urls {
		fmt.Printf("Album %d of %d:\n", albumNum+1, albumTotal)
		slug := checkUrl(_url)
		if slug == "" {
			fmt.Println("Invalid URL:", _url)
			continue
		}

		meta, err := extractMeta(_url, slug)
		if err != nil {
			panic(err)
		}

		albumFolder := filepath.Join(cfg.OutPath, sanitise(meta.Title))
		chosenFmt := chooseFmt(meta.Formats, cfg.WantedFmt)
		lowerChosenFmt := strings.ToLower(chosenFmt)

		fmt.Println(meta.Title)
		trackTotal := len(meta.Tracks)
		for trackNum, track := range meta.Tracks {
			trackNum++
			_albumFolder := albumFolder
			if meta.HasDisks {
				_albumFolder = filepath.Join(
					albumFolder, cfg.DiskNumPrefix + strconv.Itoa(track.DiskNum))
			} 
			trackPath := filepath.Join(_albumFolder, track.Fname + lowerChosenFmt)

			exists, err := fileExists(trackPath)
			if err != nil {
				handleErr(
					"Failed to check if track already exists locally.", err, false)
				continue
			}
			if exists {
				fmt.Println("Track already exists locally.")
				continue
			}

			err = makeDirs(_albumFolder)
			if err != nil {
				handleErr("Failed to make album output path.", err, false)
				continue
			}

			fmt.Printf(
				"Downloading track %d of %d: %s - %s\n", trackNum, trackTotal, track.Name,
				chosenFmt,
			)
			err = downloadTrack(trackPath, track.File + lowerChosenFmt)
			if err != nil {
				handleErr("Failed to download track.", err, false)
			}
		}
	}
}