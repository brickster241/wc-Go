package services

// WCResult struct to hold the word count results for a file
type WCResult struct {
	FileName string // Name of the file
	Lines    int    // Number of lines
	Words    int    // Number of words
	Bytes    int    // Number of bytes
	Chars    int    // Number of characters
}
