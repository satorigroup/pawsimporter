package risk

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/twinj/uuid"
	. "satori/pawsimporter/paws"
	"strings"
)

var TABLE string = GetPAWSInfo("RISK_TABLE", "LibRisksV3")
var ID string = GetPAWSInfo("RISK_ID_FIELD", "lrk_ID")
var OBJECTIVE_ID string = GetPAWSInfo("RISK_OBJ_FIELD", "lrk_RiskObjectiveID")

type Risk struct {
	DB      *gorm.DB
	Columns []Column
}

func New(db *gorm.DB, columns []Column) *Risk {
	rskSvc := &Risk{DB: db, Columns: columns}
	rskSvc.addObjectiveKey()

	return rskSvc
}

func (a *Risk) addObjectiveKey() {
	objIdColumn := Column{Import: false, Key: true, Index: -1, Name: OBJECTIVE_ID}
	a.Columns = append(a.Columns, objIdColumn)
}

func (a *Risk) Update(columnsData []Data, rowIndex int, objId string) string {
	exist, id := a.exist(columnsData, rowIndex)
	insertSql := ""
	if !exist {
		insertSql, id = a.getInsertString(columnsData)
		a.DB.Exec(insertSql)
		color.Blue("Adding a Risk which is specified in row number %d", rowIndex)
	}
	updateString, err := a.getUpdateString(columnsData)
	if err != nil {
		color.Red("No fields available for risk in row number %d", rowIndex)
		return ""
	}

	whereString, err := a.getWhereString(columnsData)
	updateResult := a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))
	if updateResult.Error != nil {
		color.Red("error occurred at row no %d while updating risk : %s", rowIndex, updateResult.Error)
		return ""
	}
	color.Green("Risk is updated at row number %d", rowIndex)
	return id
}

func (a *Risk) getInsertString(columnsData []Data) (string, string) {
	rawQuery := fmt.Sprintf("INSERT INTO %s ", TABLE)

	fieldString := ""
	valueString := ""

	for _, column := range a.Columns {
		fieldString += fmt.Sprintf("%s, ", column.Name)
		for _, columnData := range columnsData {
			if column.Index != columnData.Index {
				continue
			}
			v := a.safeSQLValue(columnData.Value)
			if column.SqlType == "bit" {
				v = a.setBitValue(v)
			}

			valueString += fmt.Sprintf("'%s', ", v)
		}
	}
	fieldString = strings.TrimRight(fieldString, ", ")
	valueString = strings.TrimRight(valueString, ", ")
	guid := uuid.NewV4().String()
	rawQuery += fmt.Sprintf("(%s, %s) VALUES ('%s', %s)", ID, fieldString, guid, valueString)

	return rawQuery, guid
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
			v := a.safeSQLValue(columnData.Value)
			if column.SqlType == "bit" {
				v = a.setBitValue(v)
			}
			rawQuery += fmt.Sprintf(" %s = '%s', ", a.safeSQLColumn(column.Name), v)
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

func (a *Risk) setBitValue(input string) string {
	input = strings.ToLower(input)
	if input == "y" || input == "yes" {
		return "1"
	}
	return "0"
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
