package everest

import (
	"github.com/csimplestring/go-left-right/primitive"
)

const (
	NumberOfBids int = 50
)

type dataType [NumberOfBids]string

type Data struct {
	*primitive.LeftRightPrimitive

	left  *dataType
	right *dataType
}

func NewData() *Data {
	var left dataType
	var right dataType

	d := &Data{
		left:  &left,
		right: &right,
	}

	d.LeftRightPrimitive = primitive.New()

	return d
}

func (d *Data) Get(index int) string {
	val := ""
	d.ApplyReadFn(d.left, d.right, func(instance interface{}) {
		i := instance.(*dataType)
		val = i[index]
	})

	return val
}

func (d *Data) Put(index int, val string) {
	d.ApplyWriteFn(d.left, d.right, func(instance interface{}) {
		i := instance.(*dataType)
		i[index] = val
	})
}
