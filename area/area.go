package area

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	. "satori/pawsimporter/paws"
	"strings"
)

var TABLE string = GetPAWSInfo("AREA_TABLE", "LibAreas")

var ID string = GetPAWSInfo("AREA_ID_FIELD", "laa_ID")

type Area struct {
	DB      *gorm.DB
	Columns []Column
}

type SqlResult struct {
	LaaArearef string `gorm:"column:laa_AreaRef"`
}

func New(db *gorm.DB, columns []Column) *Area {
	areaSvc := &Area{DB: db, Columns: columns}

	return areaSvc
}

func (a *Area) Update(columnsData []Data, rowIndex int) string {
	exist, areaId := a.exist(columnsData, rowIndex)
	if !exist {
		color.Blue("Area does not exist which is specified in row number %d", rowIndex)
		return ""
	}
	updateString, err := a.getUpdateString(columnsData)
	if err != nil {
		color.Red("No fields available for area in row number %d", rowIndex)
		return ""
	}

	whereString, err := a.getWhereString(columnsData)
	updateResult := a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))

	if updateResult.Error != nil {
		color.Red("error occurred at row no %d while updating area : %s", rowIndex, updateResult.Error)
		return ""
	}

	color.Green("Area is updated at row number %d", rowIndex)
	return areaId
}

func (a *Area) getUpdateString(columnsData []Data) (string, error) {
	rawQuery := fmt.Sprintf("Update %s SET ", TABLE)
	fieldsExist := false
	for _, column := range a.Columns {
		if column.Key {
			continue
		}
		for _, columnData := range columnsData {
			if column.Index != columnData.Index {
				continue
			}
			if !fieldsExist {
				fieldsExist = true
			}
			rawQuery += fmt.Sprintf(" %s = '%s', ", a.safeSQLColumn(column.Name), a.safeSQLValue(columnData.Value))
		}
	}
	if !fieldsExist {
		return "", errors.New("no fields exist")
	}

	rawQuery = strings.TrimRight(rawQuery, ", ")
	return rawQuery, nil
}

func (a *Area) getWhereString(columnsData []Data) (string, error) {
	refs := a.getRefValue(columnsData)
	rawQuery := ""

	if len(refs) < 1 {
		return "", errors.New("there is no reference fields")
	}

	i := 1
	for key, value := range refs {
		if i == len(refs) {
			rawQuery += fmt.Sprintf(" %s = '%s'", a.safeSQLColumn(key), a.safeSQLValue(value))

		} else {
			rawQuery += fmt.Sprintf(" %s = '%s' AND ", a.safeSQLColumn(key), a.safeSQLValue(value))

		}
		i++
	}

	return rawQuery, nil
}

func (a *Area) safeSQLValue(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Area) safeSQLColumn(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Area) getRefValue(columnsData []Data) map[string]string {
	ref := make(map[string]string)
	for _, column := range a.Columns {
		if column.Key {
			for _, columnData := range columnsData {
				if columnData.Index == column.Index {
					ref[column.Name] = columnData.Value
				}
			}
		}
	}
	return ref
}

func (a *Area) exist(columnsData []Data, rowIndex int) (bool, string) {
	whereString, err := a.getWhereString(columnsData)

	if err != nil {
		color.Red("No reference fields for area in row number %d", rowIndex)
		return false, ""
	}

	id := ""

	row := a.DB.Table(TABLE).Select(fmt.Sprintf("convert(nvarchar(36), %s) as id", ID)).Where(whereString).Row()
	row.Scan(&id)

	if id != "" {
		return true, id
	}

	return false, ""
}
