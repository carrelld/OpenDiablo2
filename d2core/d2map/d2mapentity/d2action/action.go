package d2action

import (
	"math"

	"github.com/OpenDiablo2/OpenDiablo2/d2common"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
)

type Action interface {
	AnimationMode() d2enum.PlayerAnimationMode
	OnComplete(listener func())
	Advance(elapsed float64) error //maybe?
}

type MovableEntity interface {
	Speed() float64
	GetPosition() (x, y float64)
	SetPosition(x, y float64)
	SetDirection(angle int)
}

type Move struct {
	entity  MovableEntity
	targetX float64
	targetY float64

	onCompleteFn func()
}

func NewMoveAction(m MovableEntity, x, y float64) *Move {
	move := Move{
		entity:       m,
		targetX:      x,
		targetY:      y,
		onCompleteFn: func() {},
	}
	return &move
}

func (m *Move) AnimationMode() d2enum.PlayerAnimationMode {
	return d2enum.AnimationModePlayerWalk
}

func (m *Move) OnComplete(listener func()) {
	m.onCompleteFn = listener
}

func (m *Move) Advance(elapsed float64) error {
	m.Step(elapsed)
	if m.IsAtTarget() {
		m.onCompleteFn()
	}
	return nil
}

//if m.directioner != nil {
//	angle := 359 - d2common.GetAngleBetween(
//		m.LocationX,
//		m.LocationY,
//		tx,
//		ty,
//	)
//	m.directioner(angleToDirection(float64(angle)))
//}

func angleToDirection(angle float64) int {
	degreesPerDirection := 360.0 / 64.0
	offset := 45.0 - (degreesPerDirection / 2)

	newDirection := int((angle - offset) / degreesPerDirection)

	if newDirection >= 64 {
		newDirection = newDirection - 64
	} else if newDirection < 0 {
		newDirection = 64 + newDirection
	}

	return newDirection
}

func getStepLength(speed float64, angleDegrees int, tickTime float64) (float64, float64) {
	length := tickTime * speed

	radians := (math.Pi / 180.0) * float64(angleDegrees)
	oneStepX := length * math.Cos(radians)
	oneStepY := length * math.Sin(radians)
	return oneStepX, oneStepY
}

func (m *Move) IsAtTarget() bool {
	x, y := m.entity.GetPosition()
	return math.Abs(x-m.targetX) < 0.0001 && math.Abs(y-m.targetY) < 0.0001
}

func (m *Move) Step(tickTime float64) {
	if m.IsAtTarget() {
		return
	}

	entityX, entityY := m.entity.GetPosition()
	angle := 359 - d2common.GetAngleBetween(
		entityX,
		entityY,
		m.targetX,
		m.targetY,
	)

	stepX, stepY := getStepLength(m.entity.Speed(), angle, tickTime)

	newX, _ := d2common.AdjustWithRemainder(entityX, stepX, m.targetX)
	newY, _ := d2common.AdjustWithRemainder(entityX, stepY, m.targetY)

	m.entity.SetPosition(newX, newY)
}
