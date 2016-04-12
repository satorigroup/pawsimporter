package control

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	"github.com/twinj/uuid"
	. "satori/pawsimporter/paws"
	"strings"
)

var TABLE string = GetPAWSInfo("CTR_TABLE", "LibControlsV3")
var RISKCONTROL_TABLE string = GetPAWSInfo("CTR_RSK_TABLE", "LibRiskControlLinks")
var ID string = GetPAWSInfo("CTR_ID_FIELD", "lct_Id")
var RISKCONTROL_RISK_ID string = GetPAWSInfo("L_RSK_ID", "lrcl_RiskID")
var RISKCONTROL_CTRL_ID string = GetPAWSInfo("L_CTR_ID", "lrcl_ControlID")
var RISKCONTROL_LINK_ID string = GetPAWSInfo("L_CTR_LINK_ID", "lrcl_ID")

type Control struct {
	DB      *gorm.DB
	Columns []Column
}

type ControlId struct {
	Id string `gorm:"column:id"`
}

func New(db *gorm.DB, columns []Column) *Control {
	ctrSvc := &Control{DB: db, Columns: columns}
	ctrSvc.addControlKey()
	ctrSvc.setColumnType()

	return ctrSvc
}

func (a *Control) GetControlIds(rskId string) []ControlId {
	var ctrIds []ControlId
	idColumn := fmt.Sprintf("%s = ?", RISKCONTROL_RISK_ID)
	a.DB.Table(RISKCONTROL_TABLE).Select(fmt.Sprintf("convert(nvarchar(36), %s) as id", RISKCONTROL_CTRL_ID)).Where(idColumn, rskId).Scan(&ctrIds)
	return ctrIds
}

func (a *Control) addControlKey() {
	ctrIdColumn := Column{Import: false, Key: true, Index: -1, Name: ID}
	a.Columns = append(a.Columns, ctrIdColumn)
}

func (a *Control) setControlIdvalue(ctrId string, columnsData *[]Data) {
	ctrDataColumn := Data{Index: -1, Value: ctrId}
	*columnsData = append(*columnsData, ctrDataColumn)
}
func (a *Control) InsertIfNot(columnsData []Data, riskId string) {
	whereString := a.getWhereStringForLink(columnsData)
	whereSqlString := fmt.Sprintf("LibRiskControlLinks.lrcl_RiskID = '%s' AND LibControlsV3.lct_Id = LibRiskControlLinks.lrcl_ControlID AND %s", riskId, whereString)

	type Result struct {
		Cid int
	}
	var r Result
	a.DB.Table("LibRiskControlLinks").Joins(", LibControlsV3").Select("count(*) AS cid").Where(whereSqlString).Scan(&r)
	if r.Cid == 0 {
		insertSqlString, controlId := a.getInsertString(columnsData)
		a.DB.Exec(insertSqlString)
		a.AddControlRiskLink(riskId, controlId)
	}
}

func (a *Control) getInsertString(columnsData []Data) (string, string) {
	rawQuery := fmt.Sprintf("INSERT INTO %s ", TABLE)

	fieldString := ""
	valueString := ""

	for _, column := range a.Columns {
		if !column.Import || column.Index < 0 {
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

	return rawQuery, guid
}

func (a *Control) Update(columnsData []Data, rowIndex int, rskId string) string {
	a.InsertIfNot(columnsData, rskId)
	a.DB.LogMode(true)
	ctrIds := a.GetControlIds(rskId)
	a.DB.LogMode(false)
	fmt.Println(ctrIds)
	for _, ctrId := range ctrIds {
		a.setControlIdvalue(ctrId.Id, &columnsData)
		exist, id := a.exist(columnsData, rowIndex)
		if !exist {
			continue
		}

		updateString, err := a.getUpdateString(columnsData)
		if err != nil {
			color.Red("No fields available for control in row number %d", rowIndex)
			return ""
		}

		whereString, err := a.getWhereString(columnsData)
		updateResult := a.DB.Exec(fmt.Sprintf("%s WHERE %s", updateString, whereString))
		if updateResult.Error != nil {
			color.Red("error occurred at row no %d while updating control : %s", rowIndex, updateResult.Error)
			return ""
		}
		color.Green("Control is updated at row number %d", rowIndex)
		return id
	}

	color.Blue("Control does not exist which is specified in row number %d", rowIndex)
	return ""
}

func (a *Control) getUpdateString(columnsData []Data) (string, error) {
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

func (a *Control) AddControlRiskLink(riskId, controlId string) {
	id := uuid.NewV4().String()
	sqlString := fmt.Sprintf("INSERT INTO %s ", RISKCONTROL_TABLE)
	sqlString += fmt.Sprintf(" (%s,%s,%s,%s,%s,%s) ", RISKCONTROL_LINK_ID, RISKCONTROL_RISK_ID, RISKCONTROL_CTRL_ID, "lrcl_LastModifiedBy", "lrcl_LastModifiedDate", "lrcl_ListOrder")
	sqlString += fmt.Sprintf("VALUES ('%s','%s','%s','%s','%s','%s') ", id, riskId, controlId, "5D9A9477-518A-49C3-8674-2212654514B9", "2016-04-16 12:13:50.107", "1")
	a.DB.Exec(sqlString)
}

func (a *Control) setColumnType() {
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
func (a *Control) setBitValue(input string) string {
	input = strings.ToLower(input)
	if input == "y" || input == "yes" {
		return "1"
	}
	return "0"
}

func (a *Control) getWhereStringForLink(columnsData []Data) string {
	refs := a.getRefValue(columnsData)
	rawQuery := ""

	if len(refs) < 1 {
		return ""
	}

	i := 1
	for key, value := range refs {
		columnName := fmt.Sprintf("%s.%s", TABLE, key)

		if i == len(refs) {
			rawQuery += fmt.Sprintf(" %s = '%s'", a.safeSQLColumn(columnName), a.safeSQLValue(value))

		} else {
			rawQuery += fmt.Sprintf(" %s = '%s' AND ", a.safeSQLColumn(columnName), a.safeSQLValue(value))

		}
		i++
	}

	return rawQuery
}

func (a *Control) getWhereString(columnsData []Data) (string, error) {
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

func (a *Control) safeSQLValue(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Control) safeSQLColumn(input string) string {
	output := strings.Replace(input, "'", `''`, -1)
	return output
}

func (a *Control) getRefValue(columnsData []Data) map[string]string {
	ref := make(map[string]string)
	for _, column := range a.Columns {
		if column.Key {
			for _, columnData := range columnsData {
				if columnData.Index == column.Index && column.Name != "" {
					ref[column.Name] = columnData.Value
				}
			}
		}
	}
	return ref
}

func (a *Control) exist(columnsData []Data, rowIndex int) (bool, string) {
	whereString, err := a.getWhereString(columnsData)

	if err != nil {
		color.Red("No reference fields for control in row number %d", rowIndex)
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
