package engine

type Object struct {
	name string
}

type GameObject struct {
	Object
	Active bool
	Component
	Layer int32
	Tag   uint16
}
