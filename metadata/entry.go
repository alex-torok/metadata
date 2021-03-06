package metadata

type Entry struct {
	key   string
	value int
}

type ParsedMetadataFile struct {
	entries []Entry
}
