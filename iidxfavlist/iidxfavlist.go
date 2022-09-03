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
	"time"
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
				music := iidx.findMusicByID(chart.EntryId)
				fav.Charts[i].Artist = music.Artist
				fav.Charts[i].Title = music.Title
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
	fmt.Printf("%s: rename orr modify mode favourite list\n", color.LightRed.Render("r"))
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
		_ ,line := scanInput()
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
		case "r":
			iidx.renameList()
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

func scanInput(prompt ...interface{}) (int, string) {
	fmt.Print(prompt...)
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		panic(err)
	}
	num, _ := strconv.Atoi(string(line))
	return num, strings.ToLower(string(line))
}

func scanOriginInput(prompt ...interface{}) (int, string) {
	fmt.Print(prompt...)
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		panic(err)
	}
	num, _ := strconv.Atoi(string(line))
	return num, string(line)
}

func (iidx *Iidxfavlist) createFavList() {
	now := time.Now().Format("20060102-150405")
	var spMusic, dpMusic IIDXMusicInfoDetail
	for _, m := range iidx.musicListByID {
		spMusic = m
		break
	}
	for _, m := range iidx.musicListByID {
		dpMusic = m
		break
	}
	iidx.favList = append(iidx.favList, IIDXFav{FileName:  "playlists/"+now + ".json", ChartCount: 2, Fav: []IIDXFavFile{
		{Name:"Favourite-SP-" + now, PlayStyle: "SP", Charts: []IIDXFavChart{{EntryId: spMusic.ID, Difficulty: LevelAnother, Title: spMusic.Title, Artist: spMusic.Artist}}},
		{Name:"Favourite-DP-" + now, PlayStyle: "DP", Charts: []IIDXFavChart{{EntryId: dpMusic.ID, Difficulty: LevelAnother, Title: dpMusic.Title, Artist: dpMusic.Artist}}},
	}})
	fmt.Println("create new fav list...")
}

func (iidx *Iidxfavlist) createFolder(list *[]IIDXFavFile) {
	now := time.Now().Format("20060102-150405")
	var music IIDXMusicInfoDetail
	for _, m := range iidx.musicListByID {
		music = m
		break
	}
	*list = append(*list, IIDXFavFile{Name: "Favourite-SP-" + now, PlayStyle: "SP", Charts: []IIDXFavChart{{
		EntryId: music.ID, Difficulty: LevelAnother, Title: music.Title, Artist: music.Artist,
	}}})
	fmt.Println("create new folder")
}

func (iidx *Iidxfavlist) findMusicByID(id int) IIDXMusicInfoDetail {
	music, ok := iidx.musicListByID[id]
	if !ok {
		music.ID = id
		music.Title = "unknown"
		music.Artist = "unknown"
		fmt.Println(music.ID, "not found in music list")
	}
	return music
}

func (iidx *Iidxfavlist) findInputSong(songNum int, charts []IIDXFavChart) (IIDXMusicInfoDetail, int){
	idx := -1
	for i, s := range charts {
		if songNum == s.EntryId {
			idx = i
			break
		}
	}
	return iidx.findMusicByID(songNum), idx
}

func getInputLevel(level int) string {
	if level <= 0 || level > 5 {
		level = 4
	}
	switch level {
	case 1:
		return LevelBasic
	case 2:
		return LevelNormal
	case 3:
		return LevelHyper
	case 4:
		return LevelAnother
	case 5:
		return LevelLegend
	}
	return ""
}

func (iidx *Iidxfavlist) renameList() {
	iidx.reloadFavList()

	for {
		iidx.printFavList(true, false, false)
		fileNum, _ := scanInput("input file number(default 0):")

		if len(iidx.favList) <= fileNum {
			continue
		}

		for i, song := range iidx.favList[fileNum].Fav {
			fmt.Printf("%s.%s.%s.%d(songs)\n", color.BgBlue.Render(color.FgBlack.Render(i)), song.Name, song.PlayStyle, len(song.Charts))
		}

		folderNum, _ := scanInput("input folder number(default 0):")
		if len(iidx.favList[fileNum].Fav) <= folderNum {
			continue
		}

		_, newName := scanOriginInput("input new folder name(current: "+levelColor(LevelAnother, iidx.favList[fileNum].Fav[folderNum].Name)+"):")
		fmt.Printf("%s rename to %s(Y/n):", iidx.favList[fileNum].Fav[folderNum].Name, levelColor(LevelBasic, newName))
		_, input := scanInput()
		if input == "n" {
			continue
		}
		iidx.favList[fileNum].Fav[folderNum].Name = newName
		_, newMode := scanOriginInput("input folder mode(default: "+iidx.favList[fileNum].Fav[folderNum].PlayStyle+"):")
		if len(newMode) > 0 {
			fmt.Printf("%s switch mode to %s(Y/n):", iidx.favList[fileNum].Fav[folderNum].PlayStyle, levelColor(LevelBasic, strings.ToUpper(newMode)))
			_, input = scanInput()
			if input == "n" {
				continue
			}
			iidx.favList[fileNum].Fav[folderNum].PlayStyle = strings.ToUpper(newMode)
		}

		bytes, _ := json.MarshalIndent(iidx.favList[fileNum].Fav, "", " ")
		err := ioutil.WriteFile(iidx.favList[fileNum].FileName, bytes, 0666)
		fmt.Println("rename fav list saved", err)
	}
}

func (iidx *Iidxfavlist) editFavList() {
	iidx.reloadFavList()

	for {
		iidx.printFavList(true, false, false)
		fileNum, input := scanInput("'b' to menu\ninput file number(default 0):")
		if input == "b" {
			return
		}
		if len(iidx.favList) <= fileNum {
			iidx.createFavList()
			fileNum = len(iidx.favList) - 1
		}
		list := iidx.favList[fileNum]

		for i, song := range list.Fav {
			fmt.Printf("%s.%s.%s.%d(songs)\n", color.BgBlue.Render(color.FgBlack.Render(i)), song.Name, song.PlayStyle, len(song.Charts))
		}
		folderNum, input := scanInput("'b' to select file\ninput folder number(default 0):")
		if input == "b" {
			continue
		}

		if len(list.Fav) <= folderNum {
			iidx.createFolder(&iidx.favList[fileNum].Fav)
			iidx.favList[fileNum].ChartCount++
			folderNum = len(iidx.favList[fileNum].Fav) - 1
		}
		charts := iidx.favList[fileNum].Fav[folderNum]
	Chart:
		for _, chart := range charts.Charts {
			fmt.Printf("%s.%s.%s\n", color.BgBlue.Render(color.FgBlack.Render(chart.EntryId)), levelColor(chart.Difficulty, chart.Title), chart.Artist)
		}
		songNum, _ := scanInput("'b' to select folder\ninput song id:")
		originMusic, idx := iidx.findInputSong(songNum, charts.Charts)

		if idx > -1 {
		Switch:
			existSongNum, input := scanInput("'b' to select target\n"+originMusic.Title+"."+ originMusic.Artist +" switch to id:")
			if input == "b" {
				goto Chart
			}
			if existSongNum == 0 {
				existSongNum = songNum
			}
			targetMusic := iidx.findMusicByID(existSongNum)
		Level:
			levelNum, input := scanInput("'b' to select music\n"+targetMusic.Title + "." + targetMusic.Artist + "(default: another(4)" +
				LevelBasic+"(1),"+LevelNormal+"(2),"+LevelHyper+"(3),"+LevelAnother+"(4),"+LevelLegend+"(5)")

			if input == "b" {
				goto Switch
			}

			level := getInputLevel(levelNum)

			fmt.Printf("%s switch to %s(Y/n):", originMusic.Title, levelColor(level,targetMusic.Title))
			_, input = scanInput()
			if input == "n" {
				goto Level
			}

			iidx.favList[fileNum].Fav[folderNum].Charts[idx].EntryId = targetMusic.ID
			iidx.favList[fileNum].Fav[folderNum].Charts[idx].Artist = targetMusic.Artist
			iidx.favList[fileNum].Fav[folderNum].Charts[idx].Title = targetMusic.Title
			iidx.favList[fileNum].Fav[folderNum].Charts[idx].Difficulty = level

			bytes, _ := json.MarshalIndent(iidx.favList[fileNum].Fav, "", " ")
			err := ioutil.WriteFile(iidx.favList[fileNum].FileName, bytes, 0666)
			fmt.Println("modify fav list saved", err)
		} else {
			ReLevel:
			levelNum, input := scanInput("'b' to select music\n"+originMusic.Title + "." + originMusic.Artist + "(default: another(4)" +
				LevelBasic+"(1),"+LevelNormal+"(2),"+LevelHyper+"(3),"+LevelAnother+"(4),"+LevelLegend+"(5):")

			if input == "b" {
				goto Chart
			}

			level := getInputLevel(levelNum)

			fmt.Printf("add %s(Y/n):", levelColor(level,originMusic.Title))

			_, input = scanInput()
			if input == "n" {
				goto ReLevel
			}

			iidx.favList[fileNum].Fav[folderNum].Charts = append(iidx.favList[fileNum].Fav[folderNum].Charts, IIDXFavChart{EntryId: originMusic.ID, Difficulty: level, Title: originMusic.Title, Artist: originMusic.Artist})
			iidx.favList[fileNum].ChartCount++

			bytes, _ := json.MarshalIndent(iidx.favList[fileNum].Fav, "", " ")
			err := ioutil.WriteFile(iidx.favList[fileNum].FileName, bytes, 0666)
			fmt.Println("add fav list saved", err)
		}
	}
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
