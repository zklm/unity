package engine

type Component struct {
	GameObject PPtr `field:"m_GameObject"`
}

type Behaviour struct {
	Component
	Enabled bool
}

type Transform struct {
	Component
	LocalRotation []float32
	LocalPosition []float32
	LocalScale    []float32
	Children      []PPtr
	Father        PPtr
}
