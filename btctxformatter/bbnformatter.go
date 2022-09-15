package btctxformatter

// Little helpers for now which will be usefull in tests etc.
// Probably we should not hardcode stuff either way but to define this tags
// in some separate document/config

const (
	MainTagStr string = "bbnm"

	testTagPrefix string = "bbt"

	DefautTestTagStr string = testTagPrefix + "0"
)

func MainTag() BabylonTag {
	return BabylonTag([]byte(MainTagStr))
}

func TestTag(idx uint8) BabylonTag {
	bytes := []byte(testTagPrefix)
	bytes = append(bytes, idx)
	return BabylonTag(bytes)
}
