# KHInsider-Downloader
KHInsider downloader written in Go.
![](https://i.imgur.com/4IRZAJq.png)
[Windows, Linux, macOS, and Android binaries](https://github.com/Sorrow446/KHInsider-Downloader/releases)

## Setup
|Option|Info|
| --- | --- |
|format|Track download quality. 1 = MP3, 2 = Opus, 3 = AAC/ALAC, 4 = best available / FLAC.
|outPath|Where to download to. Path will be made if it doesn't already exist.
|diskNumPrefix|Prefix for disk folders. Ex: `Disk `, `CD `.

## Usage
Args take priority over the config file.

Download a single album:   
`khi_dl_x64.exe https://downloads.khinsider.com/game-soundtracks/album/deadly-arts-g.a.s.p.-n64`

Download two albums and from two text files:   
`khi_dl_x64.exe https://downloads.khinsider.com/game-soundtracks/album/deadly-arts-g.a.s.p.-n64 https://downloads.khinsider.com/game-soundtracks/album/super-mario-galaxy-2 G:\1.txt G:\2.txt`

```
 _____ _____ _____         _   _            ____                _           _
|  |  |  |  |     |___ ___|_|_| |___ ___   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
|    -|     |-   -|   |_ -| | . | -_|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|__|__|__|__|_____|_|_|___|_|___|___|_|    |____/|___|_____|_|_|_|___|__,|___|___|_|

Usage: khi_dl_x64.exe [--format FORMAT] [--outpath OUTPATH] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --format FORMAT, -f FORMAT
                         Track download format.
                         1 = MP3
                         2 = Opus
                         3 = AAC/ALAC
                         4 = best available / FLAC [default: -1]
  --outpath OUTPATH, -o OUTPATH
                         Where to download to. Path will be made if it doesn't already exist.
  --help, -h             display this help and exit
  ```
  
  ## Thank you
  KHInsider-Downloader uses a modified version of ditashi's jsbeautifier-go.
  
  ## Disclaimer
- I will not be responsible for how you use KHInsider Downloader.    
- KHInsider brand and name is the registered trademark of its respective owner.    
- KHInsider Downloader has no partnership, sponsorship or endorsement with KHInsider.
