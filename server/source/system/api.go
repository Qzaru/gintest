package system

import (
	"context"

	sysModel "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type initApi struct{}

const initOrderApi = system.InitOrderSystem + 1

// auto run
func init() {
	system.RegisterInit(initOrderApi, &initApi{})
}

func (i *initApi) InitializerName() string {
	return sysModel.SysApi{}.TableName()
}

func (i *initApi) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	return ctx, db.AutoMigrate(&sysModel.SysApi{})
}

func (i *initApi) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&sysModel.SysApi{})
}

func (i *initApi) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, system.ErrMissingDBContext
	}
	entities := []sysModel.SysApi{
		{ApiGroup: "jwt", Method: "POST", Path: "/jwt/jsonInBlacklist", Description: "jwt加入黑名单(退出，必选)"},

		{ApiGroup: "系统用户", Method: "DELETE", Path: "/user/deleteUser", Description: "删除用户"},
		{ApiGroup: "系统用户", Method: "POST", Path: "/user/admin_register", Description: "用户注册"},
		{ApiGroup: "系统用户", Method: "POST", Path: "/user/getUserList", Description: "获取用户列表"},
		{ApiGroup: "系统用户", Method: "PUT", Path: "/user/setUserInfo", Description: "设置用户信息"},
		{ApiGroup: "系统用户", Method: "PUT", Path: "/user/setSelfInfo", Description: "设置自身信息(必选)"},
		{ApiGroup: "系统用户", Method: "GET", Path: "/user/getUserInfo", Description: "获取自身信息(必选)"},
		{ApiGroup: "系统用户", Method: "POST", Path: "/user/setUserAuthorities", Description: "设置权限组"},
		{ApiGroup: "系统用户", Method: "POST", Path: "/user/changePassword", Description: "修改密码（建议选择)"},
		{ApiGroup: "系统用户", Method: "POST", Path: "/user/setUserAuthority", Description: "修改用户角色(必选)"},
		{ApiGroup: "系统用户", Method: "POST", Path: "/user/resetPassword", Description: "重置用户密码"},
		{ApiGroup: "系统用户", Method: "PUT", Path: "/user/setSelfSetting", Description: "用户界面配置"},

		{ApiGroup: "api", Method: "POST", Path: "/api/createApi", Description: "创建api"},
		{ApiGroup: "api", Method: "POST", Path: "/api/deleteApi", Description: "删除Api"},
		{ApiGroup: "api", Method: "POST", Path: "/api/updateApi", Description: "更新Api"},
		{ApiGroup: "api", Method: "POST", Path: "/api/getApiList", Description: "获取api列表"},
		{ApiGroup: "api", Method: "POST", Path: "/api/getAllApis", Description: "获取所有api"},
		{ApiGroup: "api", Method: "POST", Path: "/api/getApiById", Description: "获取api详细信息"},
		{ApiGroup: "api", Method: "DELETE", Path: "/api/deleteApisByIds", Description: "批量删除api"},
		{ApiGroup: "api", Method: "GET", Path: "/api/syncApi", Description: "获取待同步API"},
		{ApiGroup: "api", Method: "GET", Path: "/api/getApiGroups", Description: "获取路由组"},
		{ApiGroup: "api", Method: "POST", Path: "/api/enterSyncApi", Description: "确认同步API"},
		{ApiGroup: "api", Method: "POST", Path: "/api/ignoreApi", Description: "忽略API"},

		{ApiGroup: "角色", Method: "POST", Path: "/authority/copyAuthority", Description: "拷贝角色"},
		{ApiGroup: "角色", Method: "POST", Path: "/authority/createAuthority", Description: "创建角色"},
		{ApiGroup: "角色", Method: "POST", Path: "/authority/deleteAuthority", Description: "删除角色"},
		{ApiGroup: "角色", Method: "PUT", Path: "/authority/updateAuthority", Description: "更新角色信息"},
		{ApiGroup: "角色", Method: "POST", Path: "/authority/getAuthorityList", Description: "获取角色列表"},
		{ApiGroup: "角色", Method: "POST", Path: "/authority/setDataAuthority", Description: "设置角色资源权限"},

		{ApiGroup: "casbin", Method: "POST", Path: "/casbin/updateCasbin", Description: "更改角色api权限"},
		{ApiGroup: "casbin", Method: "POST", Path: "/casbin/getPolicyPathByAuthorityId", Description: "获取权限列表"},

		{ApiGroup: "菜单", Method: "POST", Path: "/menu/addBaseMenu", Description: "新增菜单"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/getMenu", Description: "获取菜单树(必选)"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/deleteBaseMenu", Description: "删除菜单"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/updateBaseMenu", Description: "更新菜单"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/getBaseMenuById", Description: "根据id获取菜单"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/getMenuList", Description: "分页获取基础menu列表"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/getBaseMenuTree", Description: "获取用户动态路由"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/getMenuAuthority", Description: "获取指定角色menu"},
		{ApiGroup: "菜单", Method: "POST", Path: "/menu/addMenuAuthority", Description: "增加menu和角色关联关系"},

		{ApiGroup: "分片上传", Method: "GET", Path: "/fileUploadAndDownload/findFile", Description: "寻找目标文件（秒传）"},
		{ApiGroup: "分片上传", Method: "POST", Path: "/fileUploadAndDownload/breakpointContinue", Description: "断点续传"},
		{ApiGroup: "分片上传", Method: "POST", Path: "/fileUploadAndDownload/breakpointContinueFinish", Description: "断点续传完成"},
		{ApiGroup: "分片上传", Method: "POST", Path: "/fileUploadAndDownload/removeChunk", Description: "上传完成移除文件"},

		{ApiGroup: "文件上传与下载", Method: "POST", Path: "/fileUploadAndDownload/upload", Description: "文件上传（建议选择）"},
		{ApiGroup: "文件上传与下载", Method: "POST", Path: "/fileUploadAndDownload/deleteFile", Description: "删除文件"},
		{ApiGroup: "文件上传与下载", Method: "POST", Path: "/fileUploadAndDownload/editFileName", Description: "文件名或者备注编辑"},
		{ApiGroup: "文件上传与下载", Method: "POST", Path: "/fileUploadAndDownload/getFileList", Description: "获取上传文件列表"},
		{ApiGroup: "文件上传与下载", Method: "POST", Path: "/fileUploadAndDownload/importURL", Description: "导入URL"},

		{ApiGroup: "系统服务", Method: "POST", Path: "/system/getServerInfo", Description: "获取服务器信息"},
		{ApiGroup: "系统服务", Method: "POST", Path: "/system/getSystemConfig", Description: "获取配置文件内容"},
		{ApiGroup: "系统服务", Method: "POST", Path: "/system/setSystemConfig", Description: "设置配置文件内容"},

		{ApiGroup: "客户", Method: "PUT", Path: "/customer/customer", Description: "更新客户"},
		{ApiGroup: "客户", Method: "POST", Path: "/customer/customer", Description: "创建客户"},
		{ApiGroup: "客户", Method: "DELETE", Path: "/customer/customer", Description: "删除客户"},
		{ApiGroup: "客户", Method: "GET", Path: "/customer/customer", Description: "获取单一客户"},
		{ApiGroup: "客户", Method: "GET", Path: "/customer/customerList", Description: "获取客户列表"},

		{ApiGroup: "代码生成器", Method: "GET", Path: "/autoCode/getDB", Description: "获取所有数据库"},
		{ApiGroup: "代码生成器", Method: "GET", Path: "/autoCode/getTables", Description: "获取数据库表"},
		{ApiGroup: "代码生成器", Method: "POST", Path: "/autoCode/createTemp", Description: "自动化代码"},
		{ApiGroup: "代码生成器", Method: "POST", Path: "/autoCode/preview", Description: "预览自动化代码"},
		{ApiGroup: "代码生成器", Method: "GET", Path: "/autoCode/getColumn", Description: "获取所选table的所有字段"},
		{ApiGroup: "代码生成器", Method: "POST", Path: "/autoCode/installPlugin", Description: "安装插件"},
		{ApiGroup: "代码生成器", Method: "POST", Path: "/autoCode/pubPlug", Description: "打包插件"},

		{ApiGroup: "模板配置", Method: "POST", Path: "/autoCode/createPackage", Description: "配置模板"},
		{ApiGroup: "模板配置", Method: "GET", Path: "/autoCode/getTemplates", Description: "获取模板文件"},
		{ApiGroup: "模板配置", Method: "POST", Path: "/autoCode/getPackage", Description: "获取所有模板"},
		{ApiGroup: "模板配置", Method: "POST", Path: "/autoCode/delPackage", Description: "删除模板"},

		{ApiGroup: "代码生成器历史", Method: "POST", Path: "/autoCode/getMeta", Description: "获取meta信息"},
		{ApiGroup: "代码生成器历史", Method: "POST", Path: "/autoCode/rollback", Description: "回滚自动生成代码"},
		{ApiGroup: "代码生成器历史", Method: "POST", Path: "/autoCode/getSysHistory", Description: "查询回滚记录"},
		{ApiGroup: "代码生成器历史", Method: "POST", Path: "/autoCode/delSysHistory", Description: "删除回滚记录"},
		{ApiGroup: "代码生成器历史", Method: "POST", Path: "/autoCode/addFunc", Description: "增加模板方法"},

		{ApiGroup: "系统字典详情", Method: "PUT", Path: "/sysDictionaryDetail/updateSysDictionaryDetail", Description: "更新字典内容"},
		{ApiGroup: "系统字典详情", Method: "POST", Path: "/sysDictionaryDetail/createSysDictionaryDetail", Description: "新增字典内容"},
		{ApiGroup: "系统字典详情", Method: "DELETE", Path: "/sysDictionaryDetail/deleteSysDictionaryDetail", Description: "删除字典内容"},
		{ApiGroup: "系统字典详情", Method: "GET", Path: "/sysDictionaryDetail/findSysDictionaryDetail", Description: "根据ID获取字典内容"},
		{ApiGroup: "系统字典详情", Method: "GET", Path: "/sysDictionaryDetail/getSysDictionaryDetailList", Description: "获取字典内容列表"},

		{ApiGroup: "系统字典", Method: "POST", Path: "/sysDictionary/createSysDictionary", Description: "新增字典"},
		{ApiGroup: "系统字典", Method: "DELETE", Path: "/sysDictionary/deleteSysDictionary", Description: "删除字典"},
		{ApiGroup: "系统字典", Method: "PUT", Path: "/sysDictionary/updateSysDictionary", Description: "更新字典"},
		{ApiGroup: "系统字典", Method: "GET", Path: "/sysDictionary/findSysDictionary", Description: "根据ID获取字典（建议选择）"},
		{ApiGroup: "系统字典", Method: "GET", Path: "/sysDictionary/getSysDictionaryList", Description: "获取字典列表"},

		{ApiGroup: "操作记录", Method: "POST", Path: "/sysOperationRecord/createSysOperationRecord", Description: "新增操作记录"},
		{ApiGroup: "操作记录", Method: "GET", Path: "/sysOperationRecord/findSysOperationRecord", Description: "根据ID获取操作记录"},
		{ApiGroup: "操作记录", Method: "GET", Path: "/sysOperationRecord/getSysOperationRecordList", Description: "获取操作记录列表"},
		{ApiGroup: "操作记录", Method: "GET", Path: "/sysOperationRecord/getSysOperationRecordLoginById", Description: "获取对应用户登录记录列表"},
		{ApiGroup: "操作记录", Method: "DELETE", Path: "/sysOperationRecord/deleteSysOperationRecord", Description: "删除操作记录"},
		{ApiGroup: "操作记录", Method: "DELETE", Path: "/sysOperationRecord/deleteSysOperationRecordByIds", Description: "批量删除操作历史"},
		{ApiGroup: "操作记录", Method: "DELETE", Path: "/sysOperationRecord/deleteLoginRecordById", Description: "删除某个用户某时间段的登录记录"},

		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/products/:product_code", Description: "查看对应商品的基础信息"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/products/:product_code/reviews", Description: "查看对应商品的评论信息"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/products/:product_code/qa", Description: "查看对应商品的Q&A信息"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v2/favorites/skus", Description: "查看用户收藏的商品信息"},
		{ApiGroup: "商品信息", Method: "POST", Path: "/api/v2/favorites/skus/:sku_id", Description: "设置用户收藏的商品信息"},
		{ApiGroup: "商品信息", Method: "DELETE", Path: "/api/v2/favorites/skus/:sku_id", Description: "删除用户收藏的商品信息"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/products/:product_code/related", Description: "查看关联商品"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/products/:product_code/coordinates", Description: "查看关联商品搭配"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v2/history/viewed-skus", Description: "查看商品浏览记录"},
		{ApiGroup: "商品信息", Method: "POST", Path: "/api/v2/history/viewed-skus", Description: "添加商品浏览记录"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v2/cart", Description: "浏览购物车"},
		{ApiGroup: "商品信息", Method: "POST", Path: "/api/v2/cart/items", Description: "在购物车内增加商品"},
		{ApiGroup: "商品信息", Method: "PUT", Path: "/api/v2/cart/items/:sku_id", Description: "更新购物车内商品数量"},
		{ApiGroup: "商品信息", Method: "DELETE", Path: "/api/v2/cart/items/:sku_id", Description: "删除购物车内商品"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v2/shipping-addresses", Description: "获取配送地址信息"},
		{ApiGroup: "商品信息", Method: "POST", Path: "/api/v2/shipping-addresses", Description: "添加配送地址"},
		{ApiGroup: "商品信息", Method: "PUT", Path: "/api/v2/shipping-addresses/:address_id", Description: "更新配送地址"},
		{ApiGroup: "商品信息", Method: "DELETE", Path: "/api/v2/shipping-addresses/:address_id", Description: "删除配送地址"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/payments/methods", Description: "获取支付方式"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v2/orders/checkout/info", Description: "获取订单信息"},
		{ApiGroup: "商品信息", Method: "POST", Path: "/api/v2/orders/checkout/apply-coupon", Description: "使用优惠券"},
		{ApiGroup: "商品信息", Method: "DELETE", Path: "/api/v2/orders/checkout/apply-coupon", Description: "撤销优惠券"},
		{ApiGroup: "商品信息", Method: "POST", Path: "/api/v2/orders/checkout/use-points", Description: "使用积分"},
		{ApiGroup: "商品信息", Method: "DELETE", Path: "/api/v2/orders/checkout/use-points", Description: "撤销积分"},
		{ApiGroup: "商品信息", Method: "POST", Path: "/api/v2/orders", Description: "完成支付"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v2/orders", Description: "获取已下单订单信息"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v2/orders/:order_id", Description: "获取已下单订单详情"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/products/search", Description: "搜索商品"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/categories/tree", Description: "获取目录树"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/categories/:category_id", Description: "获取目录详情"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/campaigns", Description: "获取促销活动列表"},
		{ApiGroup: "商品信息", Method: "GET", Path: "/api/v1/campaigns/:campaign_id", Description: "获取促销活动详情"},

		{ApiGroup: "断点续传(插件版)", Method: "POST", Path: "/simpleUploader/upload", Description: "插件版分片上传"},
		{ApiGroup: "断点续传(插件版)", Method: "GET", Path: "/simpleUploader/checkFileMd5", Description: "文件完整度验证"},
		{ApiGroup: "断点续传(插件版)", Method: "GET", Path: "/simpleUploader/mergeFileMd5", Description: "上传完成合并文件"},

		{ApiGroup: "email", Method: "POST", Path: "/email/emailTest", Description: "发送测试邮件"},
		{ApiGroup: "email", Method: "POST", Path: "/email/sendEmail", Description: "发送邮件"},

		{ApiGroup: "按钮权限", Method: "POST", Path: "/authorityBtn/setAuthorityBtn", Description: "设置按钮权限"},
		{ApiGroup: "按钮权限", Method: "POST", Path: "/authorityBtn/getAuthorityBtn", Description: "获取已有按钮权限"},
		{ApiGroup: "按钮权限", Method: "POST", Path: "/authorityBtn/canRemoveAuthorityBtn", Description: "删除按钮"},

		{ApiGroup: "导出模板", Method: "POST", Path: "/sysExportTemplate/createSysExportTemplate", Description: "新增导出模板"},
		{ApiGroup: "导出模板", Method: "DELETE", Path: "/sysExportTemplate/deleteSysExportTemplate", Description: "删除导出模板"},
		{ApiGroup: "导出模板", Method: "DELETE", Path: "/sysExportTemplate/deleteSysExportTemplateByIds", Description: "批量删除导出模板"},
		{ApiGroup: "导出模板", Method: "PUT", Path: "/sysExportTemplate/updateSysExportTemplate", Description: "更新导出模板"},
		{ApiGroup: "导出模板", Method: "GET", Path: "/sysExportTemplate/findSysExportTemplate", Description: "根据ID获取导出模板"},
		{ApiGroup: "导出模板", Method: "GET", Path: "/sysExportTemplate/getSysExportTemplateList", Description: "获取导出模板列表"},
		{ApiGroup: "导出模板", Method: "GET", Path: "/sysExportTemplate/exportExcel", Description: "导出Excel"},
		{ApiGroup: "导出模板", Method: "GET", Path: "/sysExportTemplate/exportTemplate", Description: "下载模板"},
		{ApiGroup: "导出模板", Method: "POST", Path: "/sysExportTemplate/importExcel", Description: "导入Excel"},

		{ApiGroup: "公告", Method: "POST", Path: "/info/createInfo", Description: "新建公告"},
		{ApiGroup: "公告", Method: "DELETE", Path: "/info/deleteInfo", Description: "删除公告"},
		{ApiGroup: "公告", Method: "DELETE", Path: "/info/deleteInfoByIds", Description: "批量删除公告"},
		{ApiGroup: "公告", Method: "PUT", Path: "/info/updateInfo", Description: "更新公告"},
		{ApiGroup: "公告", Method: "GET", Path: "/info/findInfo", Description: "根据ID获取公告"},
		{ApiGroup: "公告", Method: "GET", Path: "/info/getInfoList", Description: "获取公告列表"},

		{ApiGroup: "参数管理", Method: "POST", Path: "/sysParams/createSysParams", Description: "新建参数"},
		{ApiGroup: "参数管理", Method: "DELETE", Path: "/sysParams/deleteSysParams", Description: "删除参数"},
		{ApiGroup: "参数管理", Method: "DELETE", Path: "/sysParams/deleteSysParamsByIds", Description: "批量删除参数"},
		{ApiGroup: "参数管理", Method: "PUT", Path: "/sysParams/updateSysParams", Description: "更新参数"},
		{ApiGroup: "参数管理", Method: "GET", Path: "/sysParams/findSysParams", Description: "根据ID获取参数"},
		{ApiGroup: "参数管理", Method: "GET", Path: "/sysParams/getSysParamsList", Description: "获取参数列表"},
		{ApiGroup: "参数管理", Method: "GET", Path: "/sysParams/getSysParam", Description: "获取参数列表"},
		{ApiGroup: "媒体库分类", Method: "GET", Path: "/attachmentCategory/getCategoryList", Description: "分类列表"},
		{ApiGroup: "媒体库分类", Method: "POST", Path: "/attachmentCategory/addCategory", Description: "添加/编辑分类"},
		{ApiGroup: "媒体库分类", Method: "POST", Path: "/attachmentCategory/deleteCategory", Description: "删除分类"},
	}
	if err := db.Create(&entities).Error; err != nil {
		return ctx, errors.Wrap(err, sysModel.SysApi{}.TableName()+"表数据初始化失败!")
	}
	next := context.WithValue(ctx, i.InitializerName(), entities)
	return next, nil
}

func (i *initApi) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	if errors.Is(db.Where("path = ? AND method = ?", "/authorityBtn/canRemoveAuthorityBtn", "POST").
		First(&sysModel.SysApi{}).Error, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
