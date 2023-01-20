package readline

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"
)

// fileHistory provides a history source based on a file.
type fileHistory struct {
	filename string
	list     []Item
}

// Item is the structure of an individual item in the History.list slice.
type Item struct {
	Index    int
	DateTime time.Time
	Block    string
}

func openHist(filename string) (list []Item, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return list, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var item Item
		err := json.Unmarshal(scanner.Bytes(), &item)
		if err != nil || len(item.Block) == 0 {
			continue
		}
		item.Index = len(list)
		list = append(list, item)
	}

	file.Close()
	return list, nil
}

// Write item to history file.
func (h *fileHistory) Write(s string) (int, error) {
	block := strings.TrimSpace(s)
	if block == "" {
		return 0, nil
	}

	item := Item{
		DateTime: time.Now(),
		Block:    block,
		Index:    len(h.list),
	}

	if len(h.list) == 0 || h.list[len(h.list)-1].Block != block {
		h.list = append(h.list, item)
	}

	line := struct {
		DateTime time.Time `json:"datetime"`
		Block    string    `json:"block"`
	}{
		Block:    block,
		DateTime: item.DateTime,
	}

	b, err := json.Marshal(line)
	if err != nil {
		return h.Len(), err
	}

	f, err := os.OpenFile(h.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return 0, err
	}

	_, err = f.Write(append(b, '\n'))
	f.Close()
	return h.Len(), err
}

// GetLine returns a specific line from the history file.
func (h *fileHistory) GetLine(i int) (string, error) {
	if i < 0 {
		return "", errors.New("cannot use a negative index when requesting historic commands")
	}
	if i < len(h.list) {
		return h.list[i].Block, nil
	}
	return "", errors.New("index requested greater than number of items in history")
}

// Len returns the number of items in the history file.
func (h *fileHistory) Len() int {
	return len(h.list)
}

// Dump returns the entire history file.
func (h *fileHistory) Dump() interface{} {
	return h.list
}
