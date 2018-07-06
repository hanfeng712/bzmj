package xzmj

import (
	"container/heap"
	"errorValue"
)

type MjCards struct {
	mPostion  uint32
	mCardSize uint32
	mCards    IntHeap
}

func NewMjCard() *MjCards {
	llXzmjCards := &MjCards{}
	return llXzmjCards
}

func (self *MjCards) Init() {
	self.mPostion = 0
	for i := 0; i < 3; i++ {
		for j := 1; j < 4; j++ {
			for k := 1; k < 10; k++ {
				heap.Push(&(self.mCards), j*10+k)
			}
		}
	}
}

func (self *MjCards) FaPai(card *IntHeap, size uint32) uint32 {
	var i uint32 = 0
	for i = 0; i < size; i++ {
		lCard := self.mCards.Pop()
		heap.Push(card, lCard)
	}

	return errorValue.ERET_OK
}

func (self *MjCards) MoPai(card *uint32) uint32 {
	if self.mCards.Len() == 0 {
		return errorValue.ERET_SYS_ERR
	}

	*card = (self.mCards.Pop()).(uint32)
	return errorValue.ERET_OK
}

func (self *MjCards) GetLeftCardsNum() uint32 {
	var lSize = self.mCards.Len()
	return uint32(lSize)
}

func (self *MjCards) IsEmpty() bool {
	if self.mCards.Len() <= 0 {
		return true
	}
	return false
}
