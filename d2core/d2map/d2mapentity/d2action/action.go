package d2action

import (
	"math"

	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2map/d2mapentity"

	"github.com/OpenDiablo2/OpenDiablo2/d2common"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2astar"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
)

type Action interface {
	AnimationMode() d2enum.PlayerAnimationMode
	OnComplete(listener func())
	Advance(elapsed float64) error //maybe?
}

type Move struct {
	entity  d2mapentity.MapEntity
	TargetX float64
	TargetY float64
	Speed   float64
	path    []d2astar.Pather
	done    func()
}

func NewMoveAction(entity d2mapentity.MapEntity, x, y float64) *Move {
	move := Move{
		entity:  entity,
		TargetX: x,
		TargetY: y,
	}
	return &move
}

func (m *Move) AnimationMode() d2enum.PlayerAnimationMode {
	panic("implement me")
}

func (m *Move) OnComplete(listener func()) {
	panic("implement me")
}

func (m *Move) Advance(elapsed float64) error {
	panic("implement me")
}

func (m *Move) SetPath(path []d2astar.Pather, done func()) {
	m.path = path
	m.done = done
}

func (m *Move) ClearPath() {
	m.path = nil
}

func (m *Move) SetSpeed(speed float64) {
	m.Speed = speed
}

func (m *Move) GetSpeed() float64 {
	return m.Speed
}

func (m *Move) getStepLength(tickTime float64) (float64, float64) {
	length := tickTime * m.Speed

	entityX, entityY := m.entity.GetPosition()
	angle := 359 - d2common.GetAngleBetween(
		entityX,
		entityY,
		m.TargetX,
		m.TargetY,
	)
	radians := (math.Pi / 180.0) * float64(angle)
	oneStepX := length * math.Cos(radians)
	oneStepY := length * math.Sin(radians)
	return oneStepX, oneStepY
}

func (m *Move) IsAtTarget() bool {
	x, y := m.entity.GetPosition()
	return math.Abs(x-m.TargetX) < 0.0001 && math.Abs(y-m.TargetY) < 0.0001 && len(m.path) == 0
}

func (m *Move) Step(tickTime float64) {
	if m.IsAtTarget() {
		if m.done != nil {
			m.done()
			m.done = nil
		}
		return
	}

	x, y := m.entity.GetPosition()
	stepX, stepY := m.getStepLength(tickTime)

	for {
		if d2common.AlmostEqual(x, m.TargetX, 0.0001) {
			stepX = 0
		}
		if d2common.AlmostEqual(y, m.TargetY, 0.0001) {
			stepY = 0
		}
		x, stepX = d2common.AdjustWithRemainder(x, stepX, m.TargetX)
		y, stepY = d2common.AdjustWithRemainder(y, stepY, m.TargetY)

		m.entity.subcellX = 1 + math.Mod(x, 5)
		m.entity.subcellY = 1 + math.Mod(y, 5)
		m.entity.TileX = int(x / 5)
		m.entity.TileY = int(y / 5)

		if d2common.AlmostEqual(x, m.TargetX, 0.01) && d2common.AlmostEqual(y, m.TargetY, 0.01) {
			if len(m.path) > 0 {
				m.entity.SetTarget(m.path[0].(*d2common.PathTile).X*5, m.path[0].(*d2common.PathTile).Y*5, m.done)

				if len(m.path) > 1 {
					m.path = m.path[1:]
				} else {
					m.path = []d2astar.Pather{}
				}
			} else {
				x = m.TargetX
				y = m.TargetY
				m.entity.subcellX = 1 + math.Mod(x, 5)
				m.entity.subcellY = 1 + math.Mod(y, 5)
				m.entity.TileX = int(x / 5)
				m.entity.TileY = int(y / 5)
			}
		}

		if stepX == 0 && stepY == 0 {
			break
		}

	}
}

// SetTarget sets target coordinates and changes animation based on proximity and direction
func (m *Move) SetTarget(tx, ty float64, done func()) {
	m.TargetX, m.TargetY = tx, ty
	m.done = done

	if m.entity.directioner != nil {
		x, y := m.entity.GetPosition()
		angle := 359 - d2common.GetAngleBetween(
			x,
			y,
			tx,
			ty,
		)
		m.entity.directioner(d2mapentity.angleToDirection(float64(angle)))
	}
}
