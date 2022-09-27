package btctxformatter

// Little helpers for now which will be usefull in tests etc.
// Probably we should not hardcode stuff either way but to define this tags
// in some separate document/config

const (
	mainTagPrefix string = "bbn"
	testTagPrefix string = "bbt"

	DefaultTestTagStr string = testTagPrefix + "0"
	DefaultMainTagStr string = testTagPrefix + "m"
)

func MainTag(idx uint8) BabylonTag {
	bytes := []byte(mainTagPrefix)
	bytes = append(bytes, idx)
	return BabylonTag(bytes)
}

func TestTag(idx uint8) BabylonTag {
	bytes := []byte(testTagPrefix)
	bytes = append(bytes, idx)
	return BabylonTag(bytes)
}
