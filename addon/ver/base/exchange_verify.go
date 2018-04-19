package base

import (
	"github.com/dming/lodos/addon/ver"
	"github.com/dming/lodos/utils/uuid"
	"github.com/dming/lodos/addon/seq"
	"fmt"
)

func NewExchangeVerify(_flag int, _uid uuid.UUID, _step seq.Sequence) (*exchangeVerify, error) {
	if _step <= 0 {
		return nil, nil
	}
	return &exchangeVerify {

	}, nil
}

// role as a special section of sequence system
type exchangeVerify struct {
	flag int
	uid uuid.UUID
	curSeq seq.Sequence
	maxSeq seq.Sequence
	step seq.Sequence
}

func (ex *exchangeVerify) Flag() int {
	return ex.flag
}

func (ex *exchangeVerify) UUID() uuid.UUID {
	return ex.uid
}

func (ex *exchangeVerify) MaxSeq() seq.Sequence {
	return ex.maxSeq
}

func (ex *exchangeVerify) CurSeq() seq.Sequence {
	return ex.curSeq
}

func (ex *exchangeVerify) Step() seq.Sequence {
	return ex.step
}

func (ex *exchangeVerify) SetMaxSeq(se seq.Sequence) error {
	if se < 0 {
		return fmt.Errorf("max sequen can not negative than 0")
	}
	ex.maxSeq = se
	ex.curSeq = se
	return nil
}

func (ex *exchangeVerify) CreateEntity(uid uuid.UUID, _type ver.EntityType, _fund int) (ver.Entity, error) {
GOTO://为了在并行计算里不产出相同的sequence
	seq := ex.curSeq
	ex.curSeq++
	if seq + 1 != ex.curSeq {
		goto GOTO
	}
	if ex.curSeq > ex.maxSeq {
		ex.maxSeq += ex.step
	}
	return NewEntity(uid, seq + 1, _type, _fund), nil
}

func (ex *exchangeVerify) CreateGoods(belong seq.Sequence, _type ver.GoodsType, _stackable bool) (ver.Goods, error) {
GOTO://为了在并行计算里不产出相同的sequence
	seq := ex.curSeq
	ex.curSeq++
	if seq + 1 != ex.curSeq {
		goto GOTO
	}
	if ex.curSeq > ex.maxSeq {
		ex.maxSeq += ex.step
	}
	return NewGood(belong, seq, _type, _stackable), nil
}

func (ex *exchangeVerify) ExchangeGoods(params ...ver.ExchangeUnit) error {
	//panic("implement me")
	//
	sumFund := 0
	goodsMap := make(map[int64]ver.Goods)
	for i := 0; i < len(params); i++ {
		exUnit := params[i]
		sumFund += exUnit.Fund()
		for g := 0; g < len(exUnit.Goods()); g++ {
			//todo : check if the goods is belong to entity,[use redis DB]
			//todo : return err if any goods isn't belong to the given entity
			//temp:
			if exUnit.Goods()[g].SelfSequence() < 0 { //means exchange goods out from entity
				if exUnit.Goods()[g].BelongTo() != params[i].Entity().SelfSequence() { //todo: check from db
					return fmt.Errorf("the goods[%s] is not belong to the entity[%s]", exUnit.Goods()[g], params[i].Entity())
				}
				seq := int64(exUnit.Goods()[g].SelfSequence())
				if _, ok := goodsMap[-seq]; ok { //if exist negative goods.sequence
					delete(goodsMap, -seq)		 //delete from goodsMap
				} else {
					goodsMap[seq] = exUnit.Goods()[g] //else save goods.sequence in map
				}
			}

		}
	}
	if sumFund != 0 {
		return fmt.Errorf("sum of fund is not equal to 0")
	}
	if len(goodsMap) != 0 {
		return fmt.Errorf("out.goods is not equal to in.goods")
	}
	//we have checked the validate, and we will exchange them now by updating each values
	for i := 0; i < len(params); i++ {
		exUnit := params[i]
		en := exUnit.Entity()
		goodsArr := exUnit.Goods()
		fund := exUnit.Entity().Fund() + exUnit.Fund()
		//update entity's fund [fund]
		params[i].Entity() = NewEntity(en.BelongTo(), en.SelfSequence(), en.EntityType(), fund)
		for g := 0; g < len(goodsArr); g++ {
			if goodsArr[g].SelfSequence() > 0 { //means get goods from others
				//update goods belong to [en.SelfSequence()]
				goodsArr[g] = NewGood(en.SelfSequence(), goodsArr[g].SelfSequence(), goodsArr[g].Type(), goodsArr[g].Stackable())
			}
		}
	}

	var err error = nil
	//todo: push to redis db
	//err = PushToRedis(params...)

	return err
}

func (ex *exchangeVerify) VerifyGoods(entity ver.Entity, goods []ver.Goods) error {
	for i := 0; i < len(goods); i++ {
		if goods[i].BelongTo() != entity.SelfSequence() {
			return fmt.Errorf("the goods[%s] is not belong to the entity[%s]", goods[i].SelfSequence(), entity.SelfSequence())
		}
	}
	return nil
}

func (ex *exchangeVerify) QueryGoods(entity uuid.UUID) (goods []ver.Goods) {
	panic("implement me")
}

type exchangeUnit struct {
	Entity ver.Entity
	Fund int
	Goods []ver.Goods
}