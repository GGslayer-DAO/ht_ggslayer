// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

import (
	"time"
)

const TableNameGameNft = "game_nfts"

// GameNft mapped from table <game_nfts>
type GameNft struct {
	ID          int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`                        // 主键ID
	GameID      int32     `gorm:"column:game_id;not null" json:"game_id"`                                   // 游戏id
	NftID       string    `gorm:"column:nft_id;not null" json:"nft_id"`                                     // nftid
	NftContract string    `gorm:"column:nft_contract;not null" json:"nft_contract"`                         // nft合约
	NftName     string    `gorm:"column:nft_name;not null" json:"nft_name"`                                 // nft名称
	NftDesc     string    `gorm:"column:nft_desc;not null" json:"nft_desc"`                                 // nft简介
	Image       string    `gorm:"column:image;not null" json:"image"`                                       // 图片
	URL         string    `gorm:"column:url;not null" json:"url"`                                           // 链接
	Price       float64   `gorm:"column:price;not null" json:"price"`                                       // 地板价
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`   // 创建时间
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:1970-01-01 08:00:01" json:"updated_at"` // 更新时间
}

// TableName GameNft's table name
func (*GameNft) TableName() string {
	return TableNameGameNft
}
