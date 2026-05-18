package chroniclesdk

// CensusEntry represents a count of unique players for a given class/race pair.
type CensusEntry struct {
	Class string `json:"class"`
	Race  string `json:"race"`
	Count int64  `json:"count"`
}
