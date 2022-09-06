package iidxfavlist

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
)

//New 新建一个应用, Options 动态传入配置(可选)
func New(opts ...Options) *Iidxfavlist {
	iidx := &Iidxfavlist{}

	//遍历传入的Options方法
	for _, opt := range opts {
		opt(iidx)
	}

	log.Default().SetFlags(0)

	return iidx
}

//Run 执行功能
func (iidx *Iidxfavlist) Run() {
	iidx.loadMusicList("data/info/0/video_music_list.xml")

	for {
		showHelp()
		cmd, arg := readCommandLine()
		switch cmd {
		case "e":
			iidx.editFavList()
		case "r":
			iidx.renameList()
		case "l":
			iidx.showFavList()
		case "s":
			iidx.searchFromSongList(arg)
		case "f":
			iidx.searchFromFavList(arg)
		case "q":
			return
		}
	}
}

func (iidx *Iidxfavlist) loadMusicList(filepath string) {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		dir, _ := os.Getwd()
		iidx.panic(dir, err)
	}

	var iidxMusic IIDXMusic
	if err := xml.Unmarshal(f, &iidxMusic); err != nil {
		iidx.panic(err)
	}

	iidx.musicListByID = make(map[int]IIDXMusicInfoDetail, len(iidxMusic.Music))
	for _, music := range iidxMusic.Music {
		music.Info.ID = music.ID
		iidx.musicListByID[music.ID] = music.Info
		iidx.musicList = append(iidx.musicList, music.Info)
	}

	iidx.logfn("music list loaded %d songs", len(iidxMusic.Music))
}

func (iidx *Iidxfavlist) loadFavList(favPath string) {
	fs, err := ioutil.ReadDir(favPath)
	if err != nil {
		dir, _ := os.Getwd()
		iidx.panic(dir, err)
	}

	iidx.favList = make([]IIDXFav, 0)

	var fileCount, folderCount, favCount = 0, 0, 0

	for _, file := range fs {
		if file.IsDir() {
			continue
		}

		fileName := path.Join(favPath, file.Name())

		bytes, e := ioutil.ReadFile(fileName)
		if e != nil {
			iidx.logfn("read file %s error: %v", file.Name(), e)
			continue
		}

		iidxFav := IIDXFav{FileName: fileName}

		if e := json.Unmarshal(bytes, &iidxFav.Fav); e != nil {
			iidx.logfn("read file %s error: %v, json: %v", file.Name(), e, string(bytes))
			continue
		}

		fileCount++

		for _, fav := range iidxFav.Fav {
			folderCount++
			for i, chart := range fav.Charts {
				music := iidx.findMusicByID(chart.EntryId)
				fav.Charts[i].Artist = music.Artist
				fav.Charts[i].Title = music.Title
				iidxFav.ChartCount++
				favCount++
			}
		}

		iidx.favList = append(iidx.favList, iidxFav)
	}

	iidx.logfn("fav list loaded %d files, %d folders, %d songs", fileCount, folderCount, favCount)
}

func (iidx *Iidxfavlist) randomMusic() IIDXMusicInfoDetail {
	music := IIDXMusicInfoDetail{0, "unknown", "unknown"}
	for _, m := range iidx.musicListByID {
		music = m
		break
	}
	return music
}

func (iidx *Iidxfavlist) createFavList(favPath string) {
	now := time.Now().Format("20060102-150405")
	music := iidx.randomMusic()

	iidx.favList = append(iidx.favList, IIDXFav{FileName: favPath + "/" + now + ".json", ChartCount: 2, Fav: []IIDXFavFile{
		{Name: "Favourite-SP-" + now, PlayStyle: "SP", Charts: []IIDXFavChart{{EntryId: music.ID, Difficulty: LevelAnother, Title: music.Title, Artist: music.Artist}}},
		{Name: "Favourite-DP-" + now, PlayStyle: "DP", Charts: []IIDXFavChart{{EntryId: music.ID, Difficulty: LevelAnother, Title: music.Title, Artist: music.Artist}}},
	}})

	bytes, err := json.MarshalIndent(iidx.favList[len(iidx.favList)-1].Fav, "", " ")
	if err != nil {
		iidx.panic(err)
	}
	if err := ioutil.WriteFile(iidx.favList[len(iidx.favList)-1].FileName, bytes, 0666); err != nil {
		iidx.panic(err)
	}
	iidx.logfn("create new fav list, music: %v", music)
}

func (iidx *Iidxfavlist) createFolder(fileName string, list *[]IIDXFavFile) {
	now := time.Now().Format("20060102-150405")
	music := iidx.randomMusic()

	*list = append(*list, IIDXFavFile{Name: "Favourite-SP-" + now, PlayStyle: "SP", Charts: []IIDXFavChart{{
		EntryId: music.ID, Difficulty: LevelAnother, Title: music.Title, Artist: music.Artist,
	}}})

	bytes, err := json.MarshalIndent(*list, "", " ")
	if err != nil {
		iidx.panic(err)
	}
	if err := ioutil.WriteFile(fileName, bytes, 0666); err != nil {
		iidx.panic(err)
	}
	iidx.logfn("create new folder, music: %v", music)
}

func (iidx *Iidxfavlist) findMusicByID(id int) IIDXMusicInfoDetail {
	music, ok := iidx.musicListByID[id]
	if !ok {
		music.ID = id
		music.Title = "unknown"
		music.Artist = "unknown"
		iidx.logfn("%d not found in music list", id)
	}
	return music
}

func (iidx *Iidxfavlist) findInputSong(songNum int, charts []IIDXFavChart) (IIDXMusicInfoDetail, int) {
	idx := -1
	for i, s := range charts {
		if songNum == s.EntryId {
			idx = i
			break
		}
	}
	return iidx.findMusicByID(songNum), idx
}

func (iidx *Iidxfavlist) renameList() {
	iidx.loadFavList("playlists")

	for {
		iidx.printFavList(true, false, false)
		fileNum, input := scanInput("'" + color.BgRed.Render("b") + "' to return\ninput file number(default 0)")

		if input == "b" {
			return
		}

		if len(iidx.favList) <= fileNum {
			continue
		}

		for i, song := range iidx.favList[fileNum].Fav {
			iidx.logfn("'%s'.%s.%s.%d(songs)", color.BgBlue.Render(i), song.Name, song.PlayStyle, len(song.Charts))
		}

		folderNum, _ := scanInput("input folder number(default 0)")
		if len(iidx.favList[fileNum].Fav) <= folderNum {
			continue
		}

		_, newName := scanInput("input new folder name(current: " + levelColor(LevelAnother, iidx.favList[fileNum].Fav[folderNum].Name) + ")")
		if len(newName) <= 0 {
			continue
		}
		iidx.logf("%s rename to %s(Y/n)", iidx.favList[fileNum].Fav[folderNum].Name, levelColor(LevelBeginner, newName))
		_, input = scanInput()
		if input == "n" {
			continue
		}
		iidx.favList[fileNum].Fav[folderNum].Name = newName
		_, newMode := scanInput("input folder mode(default: " + iidx.favList[fileNum].Fav[folderNum].PlayStyle + ")")
		if len(newMode) > 0 {
			iidx.logf("%s switch mode to %s(Y/n)", iidx.favList[fileNum].Fav[folderNum].PlayStyle, levelColor(LevelBeginner, strings.ToUpper(newMode)))
			_, input = scanInput()
			if input == "n" {
				continue
			}
			iidx.favList[fileNum].Fav[folderNum].PlayStyle = strings.ToUpper(newMode)
		}

		bytes, err := json.MarshalIndent(iidx.favList[fileNum].Fav, "", " ")
		if err != nil {
			iidx.panic(err)
		}
		if err := ioutil.WriteFile(iidx.favList[fileNum].FileName, bytes, 0666); err != nil {
			iidx.panic(err)
		}
		iidx.logfn("rename fav list saved")
	}
}

func (iidx *Iidxfavlist) replaceSong(music IIDXMusicInfoDetail, fileIdx, folderIdx, musicIdx int) bool {
	var songNum = 0

	for {
		songNum, _ = scanInput(music.Title + "." + music.Artist + " switch to id")
		if songNum > 0 {
			break
		}
	}

	targetMusic := iidx.findMusicByID(songNum)

	levelNum, _ := scanInput(targetMusic.Title + "." + targetMusic.Artist + "(default: " + levelColor(LevelAnother, "another(4)\n") +
		levelColor(LevelBeginner, LevelBeginner) + "(1)," + levelColor(LevelNormal, LevelNormal) + "(2)," + levelColor(LevelHyper, LevelHyper) + "(3)," + levelColor(LevelAnother, LevelAnother) + "(4)," + levelColor(LevelLegend, LevelLegend) + "(5)")
	level := getInputLevel(levelNum)

	iidx.logf("%s switch to %s(Y/n)", music.Title, levelColor(level, targetMusic.Title))
	_, input := scanInput()
	if input == "n" {
		iidx.logfn("repalce canceled")
		return false
	}

	iidx.favList[fileIdx].Fav[folderIdx].Charts[musicIdx].EntryId = targetMusic.ID
	iidx.favList[fileIdx].Fav[folderIdx].Charts[musicIdx].Artist = targetMusic.Artist
	iidx.favList[fileIdx].Fav[folderIdx].Charts[musicIdx].Title = targetMusic.Title
	iidx.favList[fileIdx].Fav[folderIdx].Charts[musicIdx].Difficulty = level

	bytes, err := json.MarshalIndent(iidx.favList[fileIdx].Fav, "", " ")
	if err != nil {
		iidx.panic(err)
	}
	if err := ioutil.WriteFile(iidx.favList[fileIdx].FileName, bytes, 0666); err != nil {
		iidx.panic(err)
	}

	iidx.logfn("modify fav list saved")
	return true
}

func (iidx *Iidxfavlist) createSong(music IIDXMusicInfoDetail, fileIdx, folderIdx int) bool {
	levelNum, _ := scanInput(music.Title + "." + music.Artist + "(default: " + levelColor(LevelAnother, "another(4)\n") +
		levelColor(LevelBeginner, LevelBeginner) + "(1)," + levelColor(LevelNormal, LevelNormal) + "(2)," + levelColor(LevelHyper, LevelHyper) + "(3)," + levelColor(LevelAnother, LevelAnother) + "(4)," + levelColor(LevelLegend, LevelLegend) + "(5)")

	level := getInputLevel(levelNum)

	iidx.logf("add %s(Y/n)", levelColor(level, music.Title))

	_, input := scanInput()
	if input == "n" {
		iidx.logfn("create canceled")
		return false
	}

	iidx.favList[fileIdx].Fav[folderIdx].Charts = append(iidx.favList[fileIdx].Fav[folderIdx].Charts, IIDXFavChart{EntryId: music.ID, Difficulty: level, Title: music.Title, Artist: music.Artist})
	iidx.favList[fileIdx].ChartCount++

	bytes, err := json.MarshalIndent(iidx.favList[fileIdx].Fav, "", " ")
	if err != nil {
		iidx.panic(err)
	}
	if err := ioutil.WriteFile(iidx.favList[fileIdx].FileName, bytes, 0666); err != nil {
		iidx.panic(err)
	}
	iidx.logfn("add fav list saved")
	return true
}

func (iidx *Iidxfavlist) editFavList() {
	iidx.loadFavList("playlists")

	for {
		iidx.printFavList(true, false, false)
		fileNum, input := scanInput("'" + color.BgRed.Render("b") + "' to return\ninput file number(default new)")
		if input == "b" {
			return
		}
		if len(iidx.favList) <= fileNum {
			continue
		}
		if len(input) <= 0 {
			iidx.createFavList("playlists")
			fileNum = len(iidx.favList) - 1
		}
		list := iidx.favList[fileNum]

		for i, song := range list.Fav {
			iidx.logfn("'%s'.%s.%s.%d(songs)", color.BgBlue.Render(i), song.Name, song.PlayStyle, len(song.Charts))
		}
		folderNum, input := scanInput("'" + color.BgRed.Render("b") + "' to menu\ninput folder number(default new)")
		if input == "b" {
			continue
		}

		if len(list.Fav) <= folderNum {
			continue
		}

		if len(input) <= 0 {
			iidx.createFolder(iidx.favList[fileNum].FileName, &iidx.favList[fileNum].Fav)
			iidx.favList[fileNum].ChartCount++
			folderNum = len(iidx.favList[fileNum].Fav) - 1
		}

		for {
			iidx.printFavSongs(fileNum, folderNum)
			songNum, input := scanInput("'" + color.BgRed.Render("b") + "' to menu\nselect or input new song id")
			if input == "b" {
				break
			}
			if len(input) <= 0 || songNum == 0 {
				continue
			}
			originMusic, idx := iidx.findInputSong(songNum, iidx.favList[fileNum].Fav[folderNum].Charts)
			var createOrEdited bool
			if idx > -1 {
				createOrEdited = iidx.replaceSong(originMusic, fileNum, folderNum, idx)
			} else {
				createOrEdited = iidx.createSong(originMusic, fileNum, folderNum)
			}
			if !createOrEdited {
				break
			}
		}

	}
}

func (iidx *Iidxfavlist) showFavList() {
	iidx.loadFavList("playlists")
	iidx.printFavList(true, true, true)
}

func (iidx *Iidxfavlist) searchFromSongList(searchExp string) {
	if len(searchExp) == 0 {
		return
	}

	musicID, _ := strconv.Atoi(searchExp)

	if music, ok := iidx.musicListByID[musicID]; ok {
		iidx.logfn("%s.%s.%s", color.FgRed.Render(musicID), music.Title, music.Artist)
	}
	for _, music := range iidx.musicList {
		if strings.Contains(strings.ToLower(music.Artist), strings.ToLower(searchExp)) {
			iidx.logfn("%d.%s.%s", music.ID, music.Title, color.FgRed.Render(music.Artist))
		}
		if strings.Contains(strings.ToLower(music.Title), strings.ToLower(searchExp)) {
			iidx.logfn("%d.%s.%s", music.ID, color.FgRed.Render(music.Title), music.Artist)
		}
	}
}

func (iidx *Iidxfavlist) searchFromFavList(searchExp string) {
	if len(searchExp) == 0 {
		return
	}

	musicID, _ := strconv.Atoi(searchExp)

	iidx.loadFavList("playlists")

	for _, fav := range iidx.favList {
		for _, song := range fav.Fav {
			for _, chart := range song.Charts {
				if chart.EntryId == musicID {
					iidx.logfn("%s:%s(%s).%d.%s.%s", fav.FileName, song.Name, song.PlayStyle, musicID, levelColor(chart.Difficulty, chart.Title), chart.Artist)
				}
				if strings.Contains(strings.ToLower(chart.Artist), strings.ToLower(searchExp)) {
					iidx.logfn("%s:%s(%s).%d.%s.%s", fav.FileName, song.Name, song.PlayStyle, chart.EntryId, levelColor(chart.Difficulty, chart.Title), chart.Artist)
				}
				if strings.Contains(strings.ToLower(chart.Title), strings.ToLower(searchExp)) {
					iidx.logfn("%s:%s(%s).%d.%s.%s", fav.FileName, song.Name, song.PlayStyle, chart.EntryId, levelColor(chart.Difficulty, chart.Title), chart.Artist)
				}
			}
		}
	}
}

func (iidx *Iidxfavlist) printFavList(showFile, showFolder, showChart bool) {
	for i, fav := range iidx.favList {
		if showFile {
			iidx.logfn("'%s'.%s, folders: %d, songs: %d", color.BgBlue.Render(i), fav.FileName, len(fav.Fav), fav.ChartCount)
		}
		for _, folder := range fav.Fav {
			if showFolder {
				iidx.logfn(" %s.%s", folder.Name, folder.PlayStyle)
			}
			for _, song := range folder.Charts {
				if showChart {
					iidx.logfn("  %s.%s", levelColor(song.Difficulty, song.Title), song.Artist)
				}
			}
		}
	}
}

func (iidx *Iidxfavlist) printFavSongs(fileIdx, folderIdx int) {
	for _, chart := range iidx.favList[fileIdx].Fav[folderIdx].Charts {
		iidx.logfn("'%s'.%s.%s", color.BgBlue.Render(chart.EntryId), levelColor(chart.Difficulty, chart.Title), chart.Artist)
	}
}
