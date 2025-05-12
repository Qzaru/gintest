package request

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

type SysOperationRecordSearch struct {
	system.SysOperationRecord
	request.PageInfo
}
type UserLoginSearch struct {
	system.UserLogin
	request.PageInfo
}
type UserLoginTimeSearch struct {
	system.UserLogin
	Start_time time.Time `json:"start_time" form:"start_time"`
	End_time   time.Time `json:"end_time" form:"end_time"`
	request.PageInfo
}
