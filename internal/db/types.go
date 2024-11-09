package db

import (
	"gorm.io/gorm"
	"math/big"
)

type DBProvider func() *gorm.DB

type Repo struct {
	DB *gorm.DB
	P  DBProvider
}

func NewRepo(p DBProvider) *Repo {
	db := p()
	return &Repo{
		DB: db,
		P:  p,
	}
}

type GasInfo struct {
	ID  int64  `gorm:"primaryKey;column:id"`        // 这会自动创建 id, created_at, updated_at, deleted_at 字段
	Gas string `gorm:"column:gas_value" json:"gas"` // 将 *big.Int 存储为字符串
}

// 添加这个方法来指定表名
func (GasInfo) TableName() string {
	return "gas_info" // 返回实际的表名
}

// 添加辅助方法来处理 *big.Int 的转换
func (g *GasInfo) SetGas(gas *big.Int) {
	if gas != nil {
		g.Gas = gas.String()
	}
}

func (g *GasInfo) GetGas() *big.Int {
	if g.Gas == "" {
		return nil
	}
	value := new(big.Int)
	value.SetString(g.Gas, 10)
	return value
}
