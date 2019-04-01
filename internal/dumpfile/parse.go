// Package dumpfile implements encoding and decoding of Edwood dump file.
//
// A dump file stores the state of Edwood so that it can be restored
// when Edwood is restarted.
package dumpfile

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const version = 1

// WindowType defines the type of window.
type WindowType int

const (
	Saved   WindowType = iota // Saved is a File and directory stored on disk
	Unsaved                   // Unsaved contains buffer that's not stored on disk
	Zerox                     // Zerox is a copy of a Saved or Unsaved window
	Exec                      // Exec is a window controlled by an outside process
)

// Content stores the state of Edwood.
type Content struct {
	CurrentDir string
	VarFont    string // Variable width font
	FixedFont  string // Fixed width font
	RowTag     string
	Columns    []Column
	Windows    []Window
}

// Column stores the state of a column in Edwood.
type Column struct {
	Position float64 // Position within the row (in percentage)
	Tag      string
}

// Window stores the state of a window in Edwood.
type Window struct {
	Type WindowType

	ColumnID int
	Q0       int
	Q1       int
	Position float64 // Position within the column (in percentage)
	Font     string  `json:",omitempty"`

	// ctl line has these but there is no point storing them:
	//ID	int	// we regenerate window IDs when loading
	//TagLen	int	// redundant
	//BodyLen int	// redundant
	//IsDir bool	// redundant

	Dirty bool
	Tag   string

	// Used for Type == Unsaved
	Body string `json:",omitempty"`

	// Used for Type == Exec
	ExecDir     string `json:",omitempty"`
	ExecCommand string `json:",omitempty"`
}

type versionedContent struct {
	Version int
	*Content
}

// Load parses the dump file and returns its content.
func Load(file string) (*Content, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return decode(bufio.NewReader(f))
}

func decode(r io.Reader) (*Content, error) {
	var vc versionedContent

	dec := json.NewDecoder(r)
	err := dec.Decode(&vc)
	if err != nil {
		return nil, err
	}
	if vc.Version != version {
		return nil, fmt.Errorf("dump file format %v; expected %v", vc.Version, version)
	}
	return vc.Content, nil
}

// Save encodes the dump file content and writes it to file.
func (c *Content) Save(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return c.encode(f)
}

func (c *Content) encode(w io.Writer) error {
	vc := versionedContent{
		Version: version,
		Content: c,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(&vc)
}

// LoadFonts gets the font names from the load file so we don't load
// fonts that we won't use.
func LoadFonts(file string) []string {
	// TODO(fhs): Maybe return two strings instead of a slice,
	// or remove this function altogether and have Edwood's main call Load
	// only once at the beginning.

	dump, err := Load(file)
	if err != nil {
		return nil
	}
	if dump.VarFont == "" || dump.FixedFont == "" {
		return nil
	}
	return []string{
		dump.VarFont,
		dump.FixedFont,
	}
}