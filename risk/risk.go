package risk

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	. "satori/pawsimporter/paws"
	"strings"
)

const (
	TABLE        = "LibRisksV3"
	ID           = "lrk_ID"
	OBJECTIVE_ID = "lrk_RiskObjectiveID"
)

type Risk struct {
	DB      *gorm.DB
	Columns []Column
}

func (a *Risk) Update(columnsData []Data, rowIndex int, objId string) string {
	objIdColumn := Column{}
	objIdColumn.Import = false
	objIdColumn.Key = true
	objIdColumn.Index = -1
	objIdColumn.Name = OBJECTIVE_ID
	a.Columns = append(a.Columns, objIdColumn)

	objDataColumn := Data{}
	objDataColumn.Index = -1
	objDataColumn.Value = objId
	columnsData = append(columnsData, objDataColumn)

	exist, id := a.exist(columnsData, rowIndex)

	if !exist {
		color.Blue("Risk does not exist which is specified in row number %d", rowIndex)
		return ""
	}
	updateString, err := a.getUpdateString(columnsData)
	if err != nil {
		color.Red("No fields available for risk in row number %d", rowIndex)
		return ""
	}

	whereString, err := a.getWhereString(columnsData)
	a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))
	color.Green("Risk is updated at row number %d", rowIndex)
	return id
}

func (a *Risk) getUpdateString(columnsData []Data) (string, error) {
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

func (a *Risk) getWhereString(columnsData []Data) (string, error) {
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

func (a *Risk) safeSQLValue(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Risk) safeSQLColumn(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Risk) getRefValue(columnsData []Data) map[string]string {
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

func (a *Risk) exist(columnsData []Data, rowIndex int) (bool, string) {
	whereString, err := a.getWhereString(columnsData)

	if err != nil {
		color.Red("No reference fields for risk in row number %d", rowIndex)
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
