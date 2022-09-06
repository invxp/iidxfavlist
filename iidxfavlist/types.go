package iidxfavlist

const (
	LevelBeginner = "beginner"
	LevelNormal   = "normal"
	LevelHyper    = "hyper"
	LevelAnother  = "another"
	LevelLegend   = "leggendaria"
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
