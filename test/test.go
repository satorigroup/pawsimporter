package test

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/twinj/uuid"
	. "satori/pawsimporter/paws"
	"strings"
)

var TABLE string = GetPAWSInfo("TST_TABLE", "LibTests")
var AREA_ID string = GetPAWSInfo("TST_AREA_ID", "lts_LibAreaID")
var RSK_ID string = GetPAWSInfo("TST_RSK_ID", "lts_RiskID")
var CTRL_ID string = GetPAWSInfo("TST_CTR_ID", "lts_ControlID")
var ID string = GetPAWSInfo("TST_FIELD_ID", "lts_ID")

type Test struct {
	DB             *gorm.DB
	Columns        []Column
	arcColumnAdded bool
}

func New(db *gorm.DB, columns []Column) *Test {
	testSvc := &Test{DB: db, Columns: columns}
	testSvc.addARCId()
	testSvc.setColumnType()

	return testSvc
}

// ARC => Area, Risk and Control
func (a *Test) addARCId() {
	ctrIdColumn := Column{Import: false, Key: true, Index: -1, Name: CTRL_ID}
	rskIdColumn := Column{Import: false, Key: true, Index: -2, Name: RSK_ID}
	areaIdColumn := Column{Import: false, Key: true, Index: -3, Name: AREA_ID}

	a.Columns = append(a.Columns, ctrIdColumn)
	a.Columns = append(a.Columns, rskIdColumn)
	a.Columns = append(a.Columns, areaIdColumn)

}

func (a *Test) Update(columnsData []Data, rowIndex int) {
	exist, _ := a.exist(columnsData, rowIndex)
	if !exist {
		color.Blue("Adding Test which does not exist which is specified in row number %d", rowIndex)
		insertSqlString, _ := a.getInsertString(columnsData)
		a.DB.Exec(insertSqlString)
	}
	updateString, err := a.getUpdateString(columnsData)
	if err != nil {
		color.Red("No fields available for test in row number %d", rowIndex)
		return
	}

	whereString, err := a.getWhereString(columnsData)
	updateResult := a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))
	if updateResult.Error != nil {
		color.Red("error occurred at row no %d while updating test: %s", rowIndex, updateResult.Error)
		return
	}
	color.Green("Test is updated at row number %d", rowIndex)
}

func (a *Test) setBitValue(input string) string {
	input = strings.ToLower(input)
	if input == "y" || input == "yes" {
		return "1"
	}
	return "0"
}

func (a *Test) getInsertString(columnsData []Data) (string, string) {
	rawQuery := fmt.Sprintf("INSERT INTO %s ", TABLE)

	fieldString := ""
	valueString := ""

	for _, column := range a.Columns {
		if !column.Import && column.Index > 0 {
			continue
		}

		fieldString += fmt.Sprintf("%s, ", column.Name)
		if column.IsRTF {
			fieldString += fmt.Sprintf("%s_RTF, ", column.Name)
		}

		for _, columnData := range columnsData {
			if column.Index != columnData.Index {
				continue
			}
			valueString += fmt.Sprintf("'%s', ", a.safeSQLValue(columnData.Value))
			if column.IsRTF {
				rtfString := fmt.Sprintf("%s %s %s", RTF_START, columnData.Value, RTF_END)
				valueString += fmt.Sprintf("'%s', ", a.safeSQLValue(rtfString))
			}
		}
	}
	fieldString = strings.TrimRight(fieldString, ", ")
	valueString = strings.TrimRight(valueString, ", ")
	guid := uuid.NewV4().String()
	rawQuery += fmt.Sprintf("(%s, %s) VALUES ('%s', %s)", ID, fieldString, guid, valueString)
	fmt.Println(rawQuery)
	return rawQuery, guid
}

func (a *Test) getUpdateString(columnsData []Data) (string, error) {
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
			rawQuery += fmt.Sprintf(" %s = '%s', ", a.safeSQLColumn(column.Name), a.safeSQLValue(columnData.Value))
			if column.IsRTF {
				rtfString := fmt.Sprintf("%s %s %s", RTF_START, columnData.Value, RTF_END)
				rawQuery += fmt.Sprintf(" %s_RTF = '%s', ", a.safeSQLColumn(column.Name), a.safeSQLValue(rtfString))
			}
		}
	}
	if !fieldsExist {
		return "", errors.New("no fields exist")
	}

	rawQuery = strings.TrimRight(rawQuery, ", ")
	return rawQuery, nil
}

func (a *Test) getWhereString(columnsData []Data) (string, error) {
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

func (a *Test) setColumnType() {
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

	for cIndex, column := range a.Columns {
		rtfColumn := fmt.Sprintf("%s_RTF", column.Name)
		for _, cInfo := range sqlInfo {
			rtfColumn := strings.ToLower(rtfColumn)
			if rtfColumn == strings.ToLower(cInfo.ColumnName) {
				a.Columns[cIndex].IsRTF = true

			}
		}
	}

}
func (a *Test) safeSQLValue(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Test) safeSQLColumn(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Test) getRefValue(columnsData []Data) map[string]string {
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

func (a *Test) exist(columnsData []Data, rowIndex int) (bool, string) {
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
