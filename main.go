package main

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"sync"
)

const boardPath = "board.txt"
const wordListPath = "words.txt"

type Words struct {
	mu       sync.RWMutex
	wordList []string
	wordMap  map[string]bool
}

func (w *Words) MarkAsFound(word string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.wordMap[word] = true
}

func (w *Words) HasPrefix(prefix string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, word := range w.wordList {
		if strings.HasPrefix(word, prefix) {
			return true
		}
	}
	return false
}

func (w *Words) Find(word string) (bool, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	v, ok := w.wordMap[word]
	return v, ok
}

func (w *Words) FoundWords() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	foundWords := make([]string, 0)

	for word, found := range w.wordMap {
		if found {
			foundWords = append(foundWords, word)
		}
	}

	return foundWords
}

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

	wordMap, err := loadWordList()
	if err != nil {
		fmt.Printf("Failed to load words: %s\n", err)
		return
	}

	wordList := make([]string, len(wordMap))
	i := 0
	for w := range wordMap {
		wordList[i] = w
		i++
	}

	board, err := loadBoard()
	if err != nil {
		fmt.Printf("Failed to load board: %s\n", err)
		return
	}

	words := &Words{
		wordList: wordList,
		wordMap:  wordMap,
	}

	findWords(words, board)

	points := 0

	foundWords := words.FoundWords()
	for _, word := range foundWords {
		if len(word) >= 8 {
			points += 11
		} else {
			points += 1 + len(word) - 4
		}
	}

	slices.SortFunc(foundWords, func(a string, b string) int {
		return len(b) - len(a)
	})

	for _, word := range foundWords {
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

func deepCopyBoard(b Board) Board {
	copyBoard := make(Board, len(b))
	for i := range b {
		copyBoard[i] = make([]Cell, len(b[i]))
		copy(copyBoard[i], b[i])
	}
	return copyBoard
}

func findWords(words *Words, board Board) {
	var wg sync.WaitGroup

	for y, row := range board {
		for x := range row {
			x, y := x, y
			wg.Add(1)
			go func(x int, y int, words *Words) {
				defer wg.Done()
				local := deepCopyBoard(board)
				scanFromPosition(words, local, &Position{x, y}, "")
			}(x, y, words)
		}
	}

	wg.Wait()
}

func scanFromPosition(words *Words, board Board, nextPos *Position, currentLetters string) {
	board.MarkUsed(nextPos)
	defer board.MarkUnused(nextPos)

	currentLetters += board.Letter(nextPos)

	if !words.HasPrefix(currentLetters) {
		return
	}

	alreadyFound, ok := words.Find(currentLetters)

	if ok && !alreadyFound {
		words.MarkAsFound(currentLetters)
	}

	neightbors := board.ValidNeighbors(nextPos)

	for _, n := range neightbors {
		scanFromPosition(words, board, &n, currentLetters)
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
