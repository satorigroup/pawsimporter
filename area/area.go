package area

import (
	"github.com/jinzhu/gorm"
	. "satori/pawsimporter/paws"
)

type Area struct {
	DB      *gorm.DB
	Columns []Column
	ref     map[string]string
}

func (a *Area) Update(data Data) {

}

func (a *Area) setRef() {
	a.ref = make(map[string]string)
	for _, column := range a.Columns {
		if column.Key {
			a.ref[column.Name] = column.Alias
		}
	}
}

func (a *Area) Exist(ref map[string]string) bool {
	exist := false

	return exist
}
