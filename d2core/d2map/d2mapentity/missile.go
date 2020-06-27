package d2mapentity

import (
	"fmt"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2data/d2datadict"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2resource"
	"github.com/OpenDiablo2/OpenDiablo2/d2core/d2asset"
)

type Missile struct {
	*AnimatedEntity
	record *d2datadict.MissileRecord
	speed  float64
}

func CreateMissile(x, y int, record *d2datadict.MissileRecord) (*Missile, error) {
	animation, err := d2asset.LoadAnimation(
		fmt.Sprintf("%s/%s.dcc", d2resource.MissileData, record.Animation.CelFileName),
		d2resource.PaletteUnits,
	)
	if err != nil {
		return nil, err
	}

	if record.Animation.HasSubLoop {
		animation.SetSubLoop(record.Animation.SubStartingFrame, record.Animation.SubEndingFrame)
	}

	animation.SetBlend(true)
	//animation.SetPlaySpeed(float64(record.Animation.AnimationSpeed))
	animation.SetPlayLoop(record.Animation.LoopAnimation)
	animation.PlayForward()
	entity := CreateAnimatedEntity(x, y, animation)

	result := &Missile{
		AnimatedEntity: entity,
		record:         record,
	}
	result.speed = float64(record.Velocity)
	return result, nil
}

func (m *Missile) SetRadians(rad float64, done func()) {
	//r := float64(m.record.Range)
	//
	//eX, eY := m.GetPosition()
	//x := eX + (r * math.Cos(rad))
	//y := eY + (r * math.Sin(rad))

	//m.SetTarget(x, y, done)
}

func (m *Missile) Advance(tickTime float64) {
	// TODO: collision detection
	//m.Step(tickTime)
	m.AnimatedEntity.Advance(tickTime)
}
