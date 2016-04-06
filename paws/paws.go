package paws

const (
	TABLE_INFO_ROW  = iota // 0
	IMPORT_INFO_ROW        // 1
	KEY_INFO_ROW           // 2
	ALIAS_INFO_ROW         // 3
	FIELD_INFO_ROW         // 4
	START_DATA_ROW         // 5
)

type Column struct {
	Name, Alias string
	Key, Import bool
	Index       int
}

type Data struct {
	Value string
	Index int
}
