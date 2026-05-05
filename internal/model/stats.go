package model

// Stats summarizes activity covered by a release record. The category
// counters mirror ChangeType buckets; FilesChanged is the union of all
// file paths touched by any included change.
type Stats struct {
	Changes      int `json:"changes"`
	Features     int `json:"features"`
	Fixes        int `json:"fixes"`
	Breaking     int `json:"breaking"`
	Deprecations int `json:"deprecations"`
	Perf         int `json:"perf"`
	Security     int `json:"security"`
	Docs         int `json:"docs"`
	Internal     int `json:"internal"`
	Contributors int `json:"contributors,omitempty"`
	FilesChanged int `json:"filesChanged"`
	Excluded     int `json:"excluded,omitempty"`
}
