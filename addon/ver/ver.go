package ver

import (
	"github.com/dming/lodos/utils/uuid"
	"github.com/dming/lodos/addon/seq"
)

type EntityType int8 //[2]byte
const (
	_ EntityType = iota
	Package
	Storage
	Mail
	AuctionShop
)

type ExchangeVerify interface {
	Flag() int
	UUID() uuid.UUID
	MaxSeq() seq.Sequence
	CurSeq() seq.Sequence
	Step() seq.Sequence
	SetMaxSeq(se seq.Sequence) error

	CreateEntity(uid uuid.UUID, _type EntityType, _fund int) (Entity, error)
	CreateGoods(uid uuid.UUID, _type GoodsType, _stackable bool) (Goods, error)
	ExchangeGoods(params ...ExchangeUnit) error
	VerifyGoods(entity Entity, goods []Goods) error
	QueryGoods(entity uuid.UUID) (goods []Goods)

	//getEntity(uuid, sequence, )
}

type RedisDBHandler interface {
	OnInit()
} 



type ExchangeUnit interface {
	Entity() Entity
	Fund() int
	Goods() []Goods

	UpdateFund(newFund int)
	UpdateGoods(newGoods []Goods)
}


/**
	goods sequence
	entity sequence
	user uuid

	goods belong to entity
	entity belong to user
 */

type GoodsType int
//type Goods [32]byte //[32]byte, as format [uuid][int64]
type Goods interface {
	BelongTo() seq.Sequence
	SelfSequence() seq.Sequence
	Type() GoodsType
	Stackable() bool
}

//type Entity [34]byte // [34]byte, as format [uuid][int64][int8]
type Entity interface {
	BelongTo() uuid.UUID
	SelfSequence() seq.Sequence
	EntityType() EntityType
	Fund() int
}