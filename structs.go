package main

type Transport struct{}

type WriteCounter struct {
	Total      int64
	TotalStr   string
	Downloaded int64
	Percentage int
	StartTime  int64
}

type Config struct {
	Urls      []string
	Format    int
	OutPath	  string
	DiskNumPrefix string
	WantedFmt string
}

type Args struct {
	Urls	[]string `arg:"positional, required"`
	Format  int      `arg:"-f" default:"-1" help:"Track download format.\n\t\t\t 1 = MP3\n\t\t\t 2 = Opus\n\t\t\t 3 = AAC/ALAC\n\t\t\t 4 = best available / FLAC"`
	OutPath string   `arg:"-o" help:"Where to download to. Path will be made if it doesn't already exist."`
}

type Track struct {
	Track	int    `json:"track"`
	Name	string `json:"name"`
	Length	string `json:"length"`
	File	string `json:"file"`
	DiskNum	int
	Fname   string
}

type Meta struct {
	Title	 string
	Formats	 []string
	HasDisks bool
	Tracks	 []*Track
}