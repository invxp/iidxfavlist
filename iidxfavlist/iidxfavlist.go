package iidxfavlist

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gookit/color"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type Iidxfavlist struct {
	musicListByID map[int]IIDXMusicInfoDetail
	musicList     []IIDXMusicInfoDetail

	favList []IIDXFav
}

type IIDXFav struct {
	Fav        []IIDXFavFile
	FileName   string
	ChartCount int
}

type IIDXFavFile struct {
	Name      string         `json:"name"`
	PlayStyle string         `json:"play_style"`
	Charts    []IIDXFavChart `json:"charts"`
}

type IIDXFavChart struct {
	EntryId    int    `json:"entry_id"`
	Difficulty string `json:"difficulty"`
	Title      string `json:"-"`
	Artist     string `json:"-"`
}

type IIDXMusic struct {
	Music []IIDXMusicInfo `xml:"music"`
}

type IIDXMusicInfo struct {
	Info IIDXMusicInfoDetail `xml:"info"`
	ID   int                 `xml:"id,attr"`
}

type IIDXMusicInfoDetail struct {
	ID     int    `xml:"-"`
	Title  string `xml:"title_name"`
	Artist string `xml:"artist_name"`
}

const (
	LevelBasic   = "basic"
	LevelNormal  = "normal"
	LevelHyper   = "hyper"
	LevelAnother = "another"
	LevelLegend  = "leggendaria"
)

//New 新建一个应用, Options 动态传入配置(可选)
func New(opts ...Options) (*Iidxfavlist, error) {
	iidx := &Iidxfavlist{}

	//遍历传入的Options方法
	for _, opt := range opts {
		opt(iidx)
	}

	return iidx, nil
}

func (iidx *Iidxfavlist) loadMusicList() {
	f, err := ioutil.ReadFile("data/info/0/video_music_list.xml")
	if err != nil {
		panic(err)
	}

	var iidxMusic IIDXMusic
	err = xml.Unmarshal(f, &iidxMusic)
	if err != nil {
		panic(err)
	}

	iidx.musicListByID = make(map[int]IIDXMusicInfoDetail, len(iidxMusic.Music))
	for _, music := range iidxMusic.Music {
		music.Info.ID = music.ID
		iidx.musicListByID[music.ID] = music.Info
		iidx.musicList = append(iidx.musicList, music.Info)
	}

	iidx.logf("music list loaded %d songs", len(iidxMusic.Music))
}

func (iidx *Iidxfavlist) reloadFavList() {
	fs, err := ioutil.ReadDir("playlists")
	if err != nil {
		panic(err)
	}

	iidx.favList = make([]IIDXFav, 0)

	for _, file := range fs {
		if file.IsDir() {
			continue
		}

		fileName := path.Join("playlists", file.Name())

		bytes, e := ioutil.ReadFile(fileName)
		if e != nil {
			iidx.logf("read file %s error: %v", file.Name(), bytes)
			continue
		}

		iidxFav := IIDXFav{FileName: fileName}

		e = json.Unmarshal(bytes, &iidxFav.Fav)
		if e != nil {
			iidx.logf("read file %s error: %v", file.Name(), bytes)
			continue
		}

		for _, fav := range iidxFav.Fav {
			for i, chart := range fav.Charts {
				if music, ok := iidx.musicListByID[chart.EntryId]; ok {
					fav.Charts[i].Artist = music.Artist
					fav.Charts[i].Title = music.Title
				} else {
					fav.Charts[i].Artist = "unknown"
					fav.Charts[i].Title = "unknown"
					iidx.logf("music not found: %v", chart.EntryId)
				}
				iidxFav.ChartCount++
			}
		}

		iidx.favList = append(iidx.favList, iidxFav)
	}
}

func (iidx *Iidxfavlist) showHelp() {
	fmt.Println("-----------beatmaniaIIDX favourite song list editor-----------")
	fmt.Println("-------------------------------------------------author: InvXp")
	fmt.Println("==========================COMMANDS============================")
	fmt.Printf("%s: edit favourite song\n", color.LightRed.Render("e"))
	fmt.Printf("%s: list favourite song\n", color.LightRed.Render("l"))
	fmt.Printf("%s: search from songlist.exp:'s {id}/{artist}/{songname}'\n", color.LightRed.Render("s"))
	fmt.Printf("%s: search from favlist.exp:'f {id}/{artist}/{songname}'\n", color.LightRed.Render("f"))

	fmt.Printf("%s: exit\n", color.LightRed.Render("q"))
}

//Run 执行功能
func (iidx *Iidxfavlist) Run() {
	iidx.loadMusicList()

	for {
		var searchExp string
		iidx.showHelp()
		line := scanInput()
		if len(line) <= 0 {
			continue
		}
		cmd := string(line[0])
		if len(line) > 2 {
			searchExp = strings.TrimSpace(line[2:])
		}
		switch cmd {
		case "e":
			iidx.editFavList()
		case "l":
			iidx.showFavList()
		case "s":
			iidx.searchFromSongList(searchExp)
		case "f":
			iidx.searchFromFavList(searchExp)
		case "q":
			return
		}
	}
}

func levelColor(diff, title string) string {
	c := color.BgDefault.Render
	b := color.FgBlack.Render
	switch diff {
	case LevelBasic:
		c = color.BgGreen.Render
	case LevelNormal:
		c = color.BgBlue.Render
	case LevelHyper:
		c = color.BgYellow.Render
	case LevelAnother:
		c = color.BgRed.Render
	case LevelLegend:
		c = color.BgMagenta.Render
	}
	return b(c(title))
}

func scanInput() string {
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		panic(err)
	}
	return string(line)
}

func (iidx *Iidxfavlist) editFavList() {
	iidx.reloadFavList()

	for {
		iidx.printFavList(true, false, false)
		fmt.Println("'b' to menu")
		fmt.Print("input file number(default 0):")
		input := scanInput()
		if strings.ToLower(input) == "b" {
			return
		}
		num, _ := strconv.Atoi(input)
		if len(iidx.favList) <= num {
			continue
		}
		list := iidx.favList[num]
	Folder:
		for i, song := range list.Fav {
			fmt.Printf("%s.%s.%s.%d(songs)\n", color.BgBlue.Render(color.FgBlack.Render(i)), song.Name, song.PlayStyle, len(song.Charts))
		}
		fmt.Println("'b' to select file")
		fmt.Print("input folder number(default 0):")
		input = scanInput()
		if strings.ToLower(input) == "b" {
			continue
		}
		num, _ = strconv.Atoi(input)
		if len(list.Fav) <= num {
			goto Folder
		}
		charts := list.Fav[num]
	Chart:
		for _, chart := range charts.Charts {
			fmt.Printf("%s.%s.%s\n", color.BgBlue.Render(color.FgBlack.Render(chart.EntryId)), levelColor(chart.Difficulty, chart.Title), chart.Artist)
		}
		fmt.Println("'b' to select folder")
		fmt.Print("input song id:")
		input = scanInput()
		if strings.ToLower(input) == "b" {
			goto Folder
		}
		num, _ = strconv.Atoi(input)

		for i, song := range charts.Charts {
			if num == song.EntryId {
			Switch:
				fmt.Println("'b' to select target")
				fmt.Printf("%s.%s switch to id:", levelColor(song.Difficulty, song.Title), song.Artist)
				input = scanInput()
				if strings.ToLower(input) == "b" {
					goto Chart
				}
				n, _ := strconv.Atoi(input)
				var music IIDXMusicInfoDetail
				var ok bool
				if music, ok = iidx.musicListByID[n]; ok {
					fmt.Printf("%d.%s.%s(Y/n):", music.ID, music.Title, music.Artist)
				} else {
					music.ID = n
					music.Title = "unknown"
					music.Artist = "unknown"
					fmt.Printf("%d not found in music list(Y/n):", music.ID)
				}
				if strings.ToLower(scanInput()) == "n" {
					goto Switch
				}
			Level:
				fmt.Println("'b' to select music")
				fmt.Printf("%d.%s.%s select level number(default: another(4))\n", n, levelColor(LevelAnother, music.Title), music.Artist)
				fmt.Printf("%s(1),%s(2),%s(3),%s(4),%s(5):", LevelBasic, LevelNormal, LevelHyper, LevelAnother, LevelLegend)
				realLevel := LevelAnother
				input = scanInput()
				if strings.ToLower(input) == "b" {
					goto Switch
				}
				n, _ = strconv.Atoi(input)
				if n <= 0 || n > 5 {
					n = 4
				}
				switch n {
				case 1:
					realLevel = LevelBasic
				case 2:
					realLevel = LevelNormal
				case 3:
					realLevel = LevelHyper
				case 4:
					realLevel = LevelAnother
				case 5:
					realLevel = LevelLegend
				default:
					realLevel = LevelAnother
				}
				fmt.Printf("%s switch to %s(Y/n):", levelColor(song.Difficulty, song.Title), levelColor(realLevel, music.Title))
				if strings.ToLower(scanInput()) == "n" {
					goto Level
				}
				charts.Charts[i].EntryId = music.ID
				charts.Charts[i].Artist = music.Artist
				charts.Charts[i].Title = music.Title
				charts.Charts[i].Difficulty = realLevel

				bytes, _ := json.MarshalIndent(list.Fav, "", " ")
				err := ioutil.WriteFile(list.FileName, bytes, 0x644)
				fmt.Println("fav list saved", err)
			}
		}

		goto Chart
	}
}

func (iidx *Iidxfavlist) createFavList() {
	//TODO
}

func (iidx *Iidxfavlist) showFavList() {
	iidx.reloadFavList()
	iidx.printFavList(true, true, true)
}

func (iidx *Iidxfavlist) searchFromSongList(searchExp string) {
	if len(searchExp) == 0 {
		return
	}

	musicID, _ := strconv.Atoi(searchExp)

	if music, ok := iidx.musicListByID[musicID]; ok {
		fmt.Printf("%s.%s.%s\n", color.FgRed.Render(musicID), music.Title, music.Artist)
	}
	for _, music := range iidx.musicList {
		if strings.Contains(strings.ToLower(music.Artist), strings.ToLower(searchExp)) {
			fmt.Printf("%d.%s.%s\n", music.ID, music.Title, color.FgRed.Render(music.Artist))
		}
		if strings.Contains(strings.ToLower(music.Title), strings.ToLower(searchExp)) {
			fmt.Printf("%d.%s.%s\n", music.ID, color.FgRed.Render(music.Title), music.Artist)
		}
	}
}

func (iidx *Iidxfavlist) searchFromFavList(searchExp string) {
	if len(searchExp) == 0 {
		return
	}

	musicID, _ := strconv.Atoi(searchExp)

	iidx.reloadFavList()

	for _, fav := range iidx.favList {
		for _, song := range fav.Fav {
			for _, chart := range song.Charts {
				if chart.EntryId == musicID {
					fmt.Printf("%s:%s(%s).%d.%s.%s\n", fav.FileName, song.Name, song.PlayStyle, musicID, levelColor(chart.Difficulty, chart.Title), chart.Artist)
				}
				if strings.Contains(strings.ToLower(chart.Artist), strings.ToLower(searchExp)) {
					fmt.Printf("%s:%s(%s).%d.%s.%s\n", fav.FileName, song.Name, song.PlayStyle, chart.EntryId, levelColor(chart.Difficulty, chart.Title), chart.Artist)
				}
				if strings.Contains(strings.ToLower(chart.Title), strings.ToLower(searchExp)) {
					fmt.Printf("%s:%s(%s).%d.%s.%s\n", fav.FileName, song.Name, song.PlayStyle, chart.EntryId, levelColor(chart.Difficulty, chart.Title), chart.Artist)
				}
			}
		}
	}
}

func (iidx *Iidxfavlist) printFavList(showFile, showFolder, showChart bool) {
	for i, fav := range iidx.favList {
		if showFile {
			fmt.Printf("%s.%s, folders: %d, songs: %d\n", color.BgHiBlue.Render(color.FgBlack.Render(i)), fav.FileName, len(fav.Fav), fav.ChartCount)
		}
		for j, folder := range fav.Fav {
			if showFolder {
				fmt.Printf("  %s.%s.%s\n", color.BgHiCyan.Render(color.FgBlack.Render(j)), folder.Name, folder.PlayStyle)
			}
			for _, song := range folder.Charts {
				if showChart {
					fmt.Printf("    %s.%s.%s\n", color.BgHiWhite.Render(color.FgBlack.Render(song.EntryId)), levelColor(song.Difficulty, song.Title), song.Artist)
				}
			}
		}
	}
}
