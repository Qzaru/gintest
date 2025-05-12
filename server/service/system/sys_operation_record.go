package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
)

//@author: [granty1](https://github.com/granty1)
//@function: CreateSysOperationRecord
//@description: 创建记录
//@param: sysOperationRecord model.SysOperationRecord
//@return: err error

type OperationRecordService struct{}

var OperationRecordServiceApp = new(OperationRecordService)

func (operationRecordService *OperationRecordService) CreateSysOperationRecord(sysOperationRecord system.SysOperationRecord) (err error) {
	err = global.GVA_DB.Create(&sysOperationRecord).Error
	return err
}

// 以下新加的
func (operationRecordService *OperationRecordService) CreateSysOperationRecordLogin(sysOperationRecordLogin system.SysOperationRecordLogin) (err error) {
	err = global.GVA_DB.Create(&sysOperationRecordLogin).Error
	return err
}

//@author: [granty1](https://github.com/granty1)
//@author: [piexlmax](https://github.com/piexlmax)
//@function: DeleteSysOperationRecordByIds
//@description: 批量删除记录
//@param: ids request.IdsReq
//@return: err error

func (operationRecordService *OperationRecordService) DeleteSysOperationRecordByIds(ids request.IdsReq) (err error) {
	err = global.GVA_DB.Delete(&[]system.SysOperationRecord{}, "id in (?)", ids.Ids).Error
	return err
}

//新加的

// @author: [granty1](https://github.com/granty1)
// @author: [piexlmax](https://github.com/piexlmax)
// @function: DeleteSysOperationRecordByIds
// @description: 删除某用户某时间段内的登录记录
// @Param     user_id  query    int true "用户id"
// @Param     start_time  query  string true "起始时间"
// @Param     end_time  query    string true "截止时间"
// @return: err error
func (operationRecordService *OperationRecordService) DeleteLoginRecordById(info systemReq.UserLoginTimeSearch) (err error) {
	layout := "2006-01-02 15:04:05.123"
	endtimeStr := info.End_time.Format(layout)
	starttimeStr := info.Start_time.Format(layout)
	err = global.GVA_DB.Where("user_id = ? AND datetime BETWEEN ? AND ?", info.UserID, starttimeStr, endtimeStr).Delete(&[]system.UserLogin{}).Error
	return err
}

//@author: [granty1](https://github.com/granty1)
//@function: DeleteSysOperationRecord
//@description: 删除操作记录
//@param: sysOperationRecord model.SysOperationRecord
//@return: err error

func (operationRecordService *OperationRecordService) DeleteSysOperationRecord(sysOperationRecord system.SysOperationRecord) (err error) {
	err = global.GVA_DB.Delete(&sysOperationRecord).Error
	return err
}

//@author: [granty1](https://github.com/granty1)
//@function: GetSysOperationRecord
//@description: 根据id获取单条操作记录
//@param: id uint
//@return: sysOperationRecord system.SysOperationRecord, err error

func (operationRecordService *OperationRecordService) GetSysOperationRecord(id uint) (sysOperationRecord system.SysOperationRecord, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sysOperationRecord).Error
	return
}

//@author: [granty1](https://github.com/granty1)
//@author: [piexlmax](https://github.com/piexlmax)
//@function: GetSysOperationRecordInfoList
//@description: 分页获取操作记录列表
//@param: info systemReq.SysOperationRecordSearch
//@return: list interface{}, total int64, err error

func (operationRecordService *OperationRecordService) GetSysOperationRecordInfoList(info systemReq.SysOperationRecordSearch) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&system.SysOperationRecord{})
	var sysOperationRecords []system.SysOperationRecord
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Method != "" {
		db = db.Where("method = ?", info.Method)
	}
	if info.Path != "" {
		db = db.Where("path LIKE ?", "%"+info.Path+"%")
	}
	if info.Status != 0 {
		db = db.Where("status = ?", info.Status)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Order("id desc").Limit(limit).Offset(offset).Preload("User").Find(&sysOperationRecords).Error
	return sysOperationRecords, total, err
}

// 新加的

// @function: GetSysOperationRecordLoginById
// @description: 获取用户登录记录列表
// @Param     user_id  query    int false "用户id"
// @Param     start_time  query  string false "起始时间"
// @Param     end_time  query    string false "截止时间"
// @return: list interface{}, total int64, err error
func (operationRecordService *OperationRecordService) GetSysOperationRecordLoginById(info systemReq.UserLoginTimeSearch) (list interface{}, total int64, err error) {
	// 创建db
	db := global.GVA_DB.Model(&system.UserLogin{})
	var userlogin []system.UserLogin
	// 如果有条件搜索 下方会自动创建搜索语句
	layout := "2006-01-02 15:04:05"
	if !info.Start_time.IsZero() && !info.End_time.IsZero() {

		endtimeStr := info.End_time.Format(layout)
		starttimeStr := info.Start_time.Format(layout)
		db = db.Where("datetime BETWEEN ? AND ?", starttimeStr, endtimeStr)
	} else if !info.Start_time.IsZero() {
		starttimeStr := info.Start_time.Format(layout)
		db = db.Where("datetime > ?", starttimeStr)
	} else if !info.End_time.IsZero() {
		endtimeStr := info.End_time.Format(layout)
		db = db.Where("datetime < ?", endtimeStr)
	}

	//endtimeStr := info.End_time.Format(layout)

	if info.UserID != 0 {
		db = db.Where("user_id = ?", info.UserID)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Order("id desc").Find(&userlogin).Error
	//将输出内容修缮一下
	type Returndata struct {
		User_id   int    `json:"user_id"`
		Logintime string `json:"logintime"`
	}
	var returndata []Returndata
	for i, _ := range userlogin {
		returndata = append(returndata, Returndata{
			User_id:   userlogin[i].UserID,
			Logintime: userlogin[i].LoginTime.Format(layout),
		})
	}
	return returndata, total, err
}
