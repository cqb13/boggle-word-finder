package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const boardPath = "board.txt"
const wordListPath = "words.txt"

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Println("missing output path")
		return
	}

	fp, err := os.Create(args[0])
	if err != nil {
		fmt.Printf("Failed to create output file: %s\n", err)
		return
	}
	defer fp.Close()

	wordList, err := loadWordList()
	if err != nil {
		fmt.Printf("Failed to load words: %s\n", err)
		return
	}

	board, err := loadBoard()
	if err != nil {
		fmt.Printf("Failed to load board: %s\n", err)
		return
	}

	foundWords := findWords(wordList, board)

	points := 0

	words := make([]string, len(foundWords))
	i := 0
	for word := range foundWords {
		if len(word) >= 8 {
			points += 11
		} else {
			points += 1 + len(word) - 4
		}

		words[i] = word
		i++
	}

	slices.SortFunc(words, func(a string, b string) int {
		return len(b) - len(a)
	})

	for _, word := range words {
		fmt.Fprintf(fp, "%s", fmt.Sprintf("%s\n", word))
	}

	fmt.Printf("Found %d words worth a total of %d points!\n", len(foundWords), points)

}

type Position struct {
	X int
	Y int
}

type Board [][]Cell

var offsets = [][]int{
	{-1, 0},
	{1, 0},
	{0, -1},
	{0, 1},
	{1, 1},
	{1, -1},
	{-1, 1},
	{-1, -1},
}

func (b Board) ValidNeighbors(pos *Position) []Position {
	positions := make([]Position, 0)

	for _, offset := range offsets {
		if pos.X+offset[0] >= 0 && pos.X+offset[0] < len(b[pos.Y]) && pos.Y+offset[1] >= 0 && pos.Y+offset[1] < len(b) {
			pos := Position{
				pos.X + offset[0],
				pos.Y + offset[1],
			}

			if b.CanUse(&pos) {
				continue
			}

			positions = append(positions, pos)
		}
	}

	return positions
}

func (b Board) CanUse(pos *Position) bool {
	return b[pos.Y][pos.X].used
}

func (b Board) MarkUsed(pos *Position) {
	b[pos.Y][pos.X].used = true
}

func (b Board) MarkUnused(pos *Position) {
	b[pos.Y][pos.X].used = false
}

func (b Board) Letter(pos *Position) string {
	return b[pos.Y][pos.X].letter
}

type Cell struct {
	letter string
	used   bool
}

func findWords(wordList map[string]bool, board Board) map[string]any {
	foundWords := make(map[string]any)

	for y, row := range board {
		for x := range row {
			scanFromPosition(wordList, board, &Position{x, y}, "", foundWords)
		}
	}

	return foundWords
}

func hasPrefix(words map[string]bool, prefix string) bool {
	for w := range words {
		if strings.HasPrefix(w, prefix) {
			return true
		}
	}
	return false
}

func scanFromPosition(wordList map[string]bool, board Board, nextPos *Position, currentLetters string, foundWords map[string]any) {
	board.MarkUsed(nextPos)
	defer board.MarkUnused(nextPos)

	currentLetters += board.Letter(nextPos)

	if !hasPrefix(wordList, currentLetters) {
		return
	}

	alreadyFound, ok := wordList[currentLetters]

	if ok && !alreadyFound {
		wordList[currentLetters] = true
		foundWords[currentLetters] = nil
	}

	neightbors := board.ValidNeighbors(nextPos)

	for _, n := range neightbors {
		scanFromPosition(wordList, board, &n, currentLetters, foundWords)
	}
}

func loadWordList() (map[string]bool, error) {
	fp, err := os.Open(wordListPath)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	words := make(map[string]bool)

	lines := strings.Split(string(bytes), "\n")
	lines = lines[:len(lines)-1]

	for _, word := range lines {
		words[word] = false
	}

	return words, nil
}

func loadBoard() (Board, error) {
	fp, err := os.Open(boardPath)
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	rows := strings.Split(string(bytes), "\n")
	rows = rows[:len(rows)-1]

	board := make(Board, len(rows))

	for y, row := range rows {
		letters := strings.Split(row, ",")

		cells := make([]Cell, len(letters))

		for x, letter := range letters {
			cells[x] = Cell{
				letter,
				false,
			}
		}

		board[y] = cells
	}

	return board, nil
}
