package system

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MockProductsService struct{}

func (m *MockProductsService) GetAddress(userID uint) (shippingAddressListResponse system.ShippingAddressListResponse, err error) {
	// 这里我们不实现真正的逻辑，因为我们主要测试的是数据库交互
	return system.ShippingAddressListResponse{}, nil
}

func TestGetAddress(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// 使用 Mock 的 db 创建 GORM DB
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true, // skip version check, more flexible
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm database", err)
	}
	global.GVA_DB = gormDB //  替换 global.GVA_DB 为 Mock 的 DB

	// Mock data
	userID := uint(1)
	expectedAddresses := []system.ShippingAddressInfo{
		{
			AddressID:    1,
			PostalCode:   "123456",
			Prefecture:   "Tokyo",
			City:         "Shibuya",
			AddressLine1: "Address 1",
			//AddressLine2: "Address 2",
			RecipientName: "John Doe",
			PhoneNumber:   "123-456-7890",
			IsDefault:     true,
		},
		{
			AddressID:    2,
			PostalCode:   "654321",
			Prefecture:   "Osaka",
			City:         "Namba",
			AddressLine1: "Address A",
			//AddressLine2: &("Address B"),
			RecipientName: "Jane Smith",
			PhoneNumber:   "098-765-4321",
			IsDefault:     false,
		},
	}

	// 构造 sqlmock 期望的 Rows
	columns := []string{"address_id", "postal_code", "prefecture", "city", "address_line1", "address_line2", "recipient_name", "phone_number", "is_default"}
	rows := sqlmock.NewRows(columns)
	for _, address := range expectedAddresses {
		rows.AddRow(
			address.AddressID,
			address.PostalCode,
			address.Prefecture,
			address.City,
			address.AddressLine1,
			address.AddressLine2,
			address.RecipientName,
			address.PhoneNumber,
			address.IsDefault,
		)
	}

	// 预期 SQL 查询 (忽略 SQL 语句)
	mock.ExpectQuery(".*").WillReturnRows(rows)

	// 创建 ProductsService 实例
	productsService := &ProductsService{} //  使用真实的 ProductsService

	// 调用 GetAddress 函数
	shippingAddressListResponse, err := productsService.GetAddress(userID)

	// 断言结果
	assert.NoError(t, err)
	assert.NotNil(t, shippingAddressListResponse)
	assert.Equal(t, len(expectedAddresses), len(shippingAddressListResponse.Addresses))

	// 检查切片是否为空，避免 index out of range 错误
	if len(shippingAddressListResponse.Addresses) > 0 {
		assert.Equal(t, expectedAddresses[0], shippingAddressListResponse.Addresses[0])
		if len(shippingAddressListResponse.Addresses) > 1 {
			assert.Equal(t, expectedAddresses[1], shippingAddressListResponse.Addresses[1])
		}
	}

	// 确保所有预期都被执行
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
