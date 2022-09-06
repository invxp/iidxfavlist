package iidxfavlist

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

func showHelp() {
	fmt.Println("-----------beatmaniaIIDX favourite song list editor-----------")
	fmt.Println("-------------------------------------------------author: InvXp")
	fmt.Println("==========================COMMANDS============================")

	fmt.Printf("%s: edit favourite song\n", color.LightRed.Render("e"))
	fmt.Printf("%s: list favourite song\n", color.LightRed.Render("l"))
	fmt.Printf("%s: rename or modify mode favourite list\n", color.LightRed.Render("r"))
	fmt.Printf("%s: search from songlist.exp:'s {id}/{artist}/{songname}'\n", color.LightRed.Render("s"))
	fmt.Printf("%s: search from favlist.exp:'f {id}/{artist}/{songname}'\n", color.LightRed.Render("f"))

	fmt.Printf("%s: exit\n", color.LightRed.Render("q"))
}

func levelColor(diff, title string) string {
	c := color.BgDefault.Render
	b := color.FgBlack.Render
	switch diff {
	case LevelBeginner:
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
	fmt.Print(":->")
	line, _, err := bufio.NewReader(os.Stdin).ReadLine()
	if err != nil {
		panic(err)
	}
	num, _ := strconv.Atoi(string(line))
	return num, string(line)
}

func readCommandLine() (cmd string, arg string) {
	_, line := scanInput()
	if len(line) <= 0 {
		return
	}
	cmd = string(line[0])
	if len(line) > 2 {
		arg = strings.TrimSpace(line[2:])
	}
	return
}

func getInputLevel(level int) string {
	if level <= 0 || level > 5 {
		level = 4
	}
	switch level {
	case 1:
		return LevelBeginner
	case 2:
		return LevelNormal
	case 3:
		return LevelHyper
	case 4:
		return LevelAnother
	case 5:
		return LevelLegend
	}
	return LevelAnother
}

func deleteSlice[T any](index int, slice []T) []T {
	if len(slice) == 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}
