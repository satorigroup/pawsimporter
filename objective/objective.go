package objective

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	. "satori/pawsimporter/paws"
	"strings"
)

const (
	TABLE = "LibRiskObjectivesV3"
	ID    = "lro_ID"
)

type Objective struct {
	DB      *gorm.DB
	Columns []Column
}

func (a *Objective) Update(columnsData []Data, rowIndex int) string {
	exist, objId := a.exist(columnsData, rowIndex)
	if !exist {
		color.Red("Objective does not exist which is specified in row number %d", rowIndex)
		return ""
	}
	updateString, err := a.getUpdateString(columnsData)
	if err != nil {
		color.Blue("No fields available for objective in row number %d", rowIndex)
		return objId
	}

	whereString, err := a.getWhereString(columnsData)
	a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))
	color.Green("Objective is updated at row number %d", rowIndex)
	return objId
}

func (a *Objective) getUpdateString(columnsData []Data) (string, error) {
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

func (a *Objective) getWhereString(columnsData []Data) (string, error) {
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

func (a *Objective) safeSQLValue(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Objective) safeSQLColumn(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Objective) getRefValue(columnsData []Data) map[string]string {
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

func (a *Objective) exist(columnsData []Data, rowIndex int) (bool, string) {
	whereString, err := a.getWhereString(columnsData)

	if err != nil {
		color.Red("No reference fields for objective in row number %d", rowIndex)
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
