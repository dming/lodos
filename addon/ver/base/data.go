package base

import (
	"github.com/dming/lodos/utils/uuid"
	"github.com/dming/lodos/addon/seq"
	"github.com/dming/lodos/addon/ver"
)

func NewGood(belong seq.Sequence, _seq seq.Sequence, _type ver.GoodsType, _stackable bool) ver.Goods {
	//todo : format [uuid][selfSequence] to [32]byte
	//panic("implement me")
	return &goods{
		belongTo :    belong,
		selfSequence: _seq,
		gType:        _type,
		stackable:    _stackable,
	}
}

type goods struct {
	belongTo     seq.Sequence //entitySequence
	selfSequence seq.Sequence
	gType        ver.GoodsType
	stackable    bool
}

func (g *goods) BelongTo() seq.Sequence {
	return g.belongTo
}

func (g *goods) SelfSequence() seq.Sequence {
	return g.selfSequence
}
func (g *goods) Type() ver.GoodsType {
	return g.gType
}
func (g *goods) Stackable() bool {
	return g.stackable
}


func NewEntity(belong uuid.UUID, _seq seq.Sequence, _Type ver.EntityType, _fund int) ver.Entity {
	//todo : format [uuid][selfSequence][eType] to [34]byte
	return &entity{
		belongTo :    belong,
		selfSequence: _seq,
		entityType:   _Type,
		fund:         _fund,
	}
}

type entity struct {
	belongTo     uuid.UUID //user uuid
	selfSequence seq.Sequence
	entityType   ver.EntityType
	fund         int
}

func (e *entity) BelongTo() uuid.UUID {
	return e.belongTo
}

func (e *entity) Fund() int {
	return e.fund
}

func (e *entity) SelfSequence() seq.Sequence {
	return seq.Sequence(0)
}
func (e *entity) EntityType() ver.EntityType {
	return ver.EntityType(0)
}




