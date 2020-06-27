package d2mapentity

import (
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity/d2action"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2render"
)

type MapEntity interface {
	Render(target d2render.Surface)
	Advance(tickTime float64)
	SetPosition(x, y float64)
	GetPosition() (float64, float64)
	Name() string
}

// mapEntity represents an entity on the map that can be animated
type mapEntity struct {
	weaponClass string

	// new coord system to rule them all (subtiles)
	nX, nY float64

	Action d2action.Action
}

func (m *mapEntity) SetPosition(x, y float64) {
	m.nX = x
	m.nY = y
}

// createMapEntity creates an instance of mapEntity
func createMapEntity(x, y int) mapEntity {
	locX, locY := float64(x), float64(y)
	return mapEntity{
		nX: locX,
		nY: locY,
	}
}

func (m *mapEntity) GetPosition() (float64, float64) {
	return m.nX, m.nY
}

func (m *mapEntity) Name() string {
	return ""
}
