package objective

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/twinj/uuid"
	. "satori/pawsimporter/paws"
	"strings"
)

var TABLE string = GetPAWSInfo("OBJ_TABLE", "LibRiskObjectivesV3")
var ID string = GetPAWSInfo("OBJ_ID_FIELD", "lro_ID")

type Objective struct {
	DB      *gorm.DB
	Columns []Column
}

func New(db *gorm.DB, columns []Column) *Objective {
	objSvc := &Objective{DB: db, Columns: columns}
	objSvc.setColumnType()
	return objSvc
}

func (a *Objective) Update(columnsData []Data, rowIndex int) string {
	exist, objId := a.exist(columnsData, rowIndex)
	insertSQL := ""
	if !exist {
		insertSQL, objId = a.getInsertString(columnsData)
		a.DB.Exec(insertSQL)
		color.Red("Inserting new Objective which does not exist in row number %d", rowIndex)
	}
	updateString, err := a.getUpdateString(columnsData)
	if err != nil {
		color.Blue("No fields available for objective in row number %d", rowIndex)
		return objId
	}

	whereString, err := a.getWhereString(columnsData)
	updateResult := a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))
	if updateResult.Error != nil {
		color.Red("error occurred at row no %d while updating objective : %s", rowIndex, updateResult.Error)
		return ""
	}

	color.Green("Objective is updated at row number %d", rowIndex)
	return objId
}

func (a *Objective) setColumnType() {
	type SqlColumnInfo struct {
		ColumnName string `gorm:"column:COLUMN_NAME"`
		DataType   string `gorm:"column:DATA_TYPE"`
	}
	var sqlInfo []SqlColumnInfo
	a.DB.Table("INFORMATION_SCHEMA.COLUMNS").Select("COLUMN_NAME, DATA_TYPE").Where("TABLE_NAME = ?", TABLE).Scan(&sqlInfo)
	for _, cInfo := range sqlInfo {
		for cIndex, column := range a.Columns {
			if strings.ToLower(column.Name) == strings.ToLower(cInfo.ColumnName) {
				a.Columns[cIndex].SqlType = cInfo.DataType

			}
		}
	}
}
func (a *Objective) setBitValue(input string) string {
	input = strings.ToLower(input)
	if input == "y" || input == "yes" {
		return "1"
	}
	return "0"
}
func (a *Objective) getInsertString(columnsData []Data) (string, string) {
	rawQuery := fmt.Sprintf("INSERT INTO %s ", TABLE)

	fieldString := ""
	valueString := ""

	for _, column := range a.Columns {
		if !column.Import {
			continue
		}
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

func (a *Objective) getUpdateString(columnsData []Data) (string, error) {
	rawQuery := fmt.Sprintf("Update %s SET ", TABLE)
	fieldsExist := false
	for _, column := range a.Columns {
		if column.Key || !column.Import {
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
