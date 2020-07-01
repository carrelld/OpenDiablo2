package d2asset

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2data"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2data/d2datadict"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2interface"
)

// Composite is a composite entity animation
type Composite struct {
	object      *d2datadict.ObjectLookupRecord
	palettePath string
	mode        *compositeMode
}

// CreateComposite creates a Composite from a given ObjectLookupRecord and palettePath.
func CreateComposite(object *d2datadict.ObjectLookupRecord, palettePath string) *Composite {
	return &Composite{object: object, palettePath: palettePath}
}

// Advance moves the composite animation forward for a given elapsed time in nanoseconds.
func (c *Composite) Advance(elapsed float64) error {
	if c.mode == nil {
		return nil
	}

	c.mode.lastFrameTime += elapsed
	framesToAdd := int(c.mode.lastFrameTime / c.mode.animationSpeed)
	c.mode.lastFrameTime -= float64(framesToAdd) * c.mode.animationSpeed
	c.mode.frameIndex += framesToAdd
	c.mode.playedCount += c.mode.frameIndex / c.mode.frameCount
	c.mode.frameIndex %= c.mode.frameCount

	for _, layer := range c.mode.layers {
		if layer != nil {
			if err := layer.Advance(elapsed); err != nil {
				return err
			}
		}
	}

	return nil
}

// Render performs drawing of the Composite on the rendered d2interface.Surface.
func (c *Composite) Render(target d2interface.Surface) error {
	if c.mode == nil {
		return nil
	}

	for _, layerIndex := range c.mode.drawOrder[c.mode.frameIndex] {
		layer := c.mode.layers[layerIndex]
		if layer != nil {
			if err := layer.RenderFromOrigin(target); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetAnimationMode returns the animation mode the Composite should render with.
func (c Composite) GetAnimationMode() string {
	return c.mode.animationMode
}

// SetMode sets the Composite's animation mode weapon class and direction
func (c *Composite) SetMode(animationMode, weaponClass string, direction int) error {
	if c.mode != nil && c.mode.animationMode == animationMode && c.mode.weaponClass == weaponClass && c.mode.direction == direction {
		return nil
	}

	mode, err := c.createMode(animationMode, weaponClass, direction)
	if err != nil {
		return err
	}

	c.resetPlayedCount()
	c.mode = mode

	return nil
}

// SetSpeed sets the speed at which the Composite's animation should advance through its frames
func (c *Composite) SetSpeed(speed int) {
	c.mode.animationSpeed = defaultAnimationNumerator / ((float64(speed) * 25.0) / 256.0)
	for layerIdx := range c.mode.layers {
		layer := c.mode.layers[layerIdx]
		if layer != nil {
			layer.SetPlaySpeed(c.mode.animationSpeed)
		}
	}
}

// GetDirectionCount returns the Composites number of available animated directions
func (c *Composite) GetDirectionCount() int {
	if c.mode == nil {
		return 0
	}

	return c.mode.directionCount
}

// GetPlayedCount returns the number of times the current animation mode has completed all its distinct frames
func (c *Composite) GetPlayedCount() int {
	if c.mode == nil {
		return 0
	}

	return c.mode.playedCount
}

func (c *Composite) resetPlayedCount() {
	if c.mode != nil {
		c.mode.playedCount = 0
	}
}

type compositeMode struct {
	animationMode  string
	weaponClass    string
	direction      int
	directionCount int
	playedCount    int

	layers    []*Animation
	drawOrder [][]d2enum.CompositeType

	frameCount     int
	frameIndex     int
	animationSpeed float64
	lastFrameTime  float64
}

const defaultAnimationNumerator = 1.0 // TODO figure out a good name for this

func (c *Composite) createMode(animationMode, weaponClass string, direction int) (*compositeMode, error) {
	cofPath := fmt.Sprintf("%s/%s/COF/%s%s%s.COF", c.object.Base, c.object.Token, c.object.Token, animationMode, weaponClass)
	if exists, _ := FileExists(cofPath); !exists {
		return nil, errors.New("composite not found")
	}

	cof, err := loadCOF(cofPath)
	if err != nil {
		return nil, err
	}

	offset := (64 / cof.NumberOfDirections) / 2
	entityDirection := int(math.Trunc((float64(direction+offset)-64.0)*(-float64(cof.NumberOfDirections)/-64.0) + float64(cof.NumberOfDirections)))

	if entityDirection >= cof.NumberOfDirections {
		entityDirection = 0
	}

	animationKey := strings.ToLower(c.object.Token + animationMode + weaponClass)

	animationData := d2data.AnimationData[animationKey]
	if len(animationData) == 0 {
		return nil, errors.New("could not find animation data")
	}

	mode := &compositeMode{
		animationMode:  animationMode,
		weaponClass:    weaponClass,
		direction:      entityDirection,
		directionCount: cof.NumberOfDirections,
		layers:         make([]*Animation, d2enum.CompositeTypeMax),
		frameCount:     animationData[0].FramesPerDirection,
		animationSpeed: defaultAnimationNumerator / ((float64(animationData[0].AnimationSpeed) * 25.0) / 256.0),
	}

	mode.drawOrder = make([][]d2enum.CompositeType, mode.frameCount)
	for frame := 0; frame < mode.frameCount; frame++ {
		mode.drawOrder[frame] = cof.Priority[mode.direction][frame]
	}

	for _, cofLayer := range cof.CofLayers {
		layerKey, layerValue, err := compositeTypeLayerKeyValue(cofLayer.Type, c.object)
		if err != nil {
			return nil, err
		}

		blend := false
		transparency := 255

		if cofLayer.Transparent {
			switch cofLayer.DrawEffect {
			case d2enum.DrawEffectPctTransparency25:
				transparency = 64
			case d2enum.DrawEffectPctTransparency50:
				transparency = 128
			case d2enum.DrawEffectPctTransparency75:
				transparency = 192
			case d2enum.DrawEffectModulate:
				blend = true
			}
		}

		layer, err := loadCompositeLayer(c.object, layerKey, layerValue, animationMode, weaponClass, c.palettePath, transparency)
		if err == nil {
			layer.SetPlaySpeed(mode.animationSpeed)
			layer.PlayForward()
			layer.SetBlend(blend)

			if err := layer.SetDirection(direction); err != nil {
				return nil, err
			}

			mode.layers[cofLayer.Type] = layer
		}
	}

	return mode, nil
}

func compositeTypeLayerKeyValue(ct d2enum.CompositeType, object *d2datadict.ObjectLookupRecord) (layerKey, layerValue string, err error) {
	switch ct {
	case d2enum.CompositeTypeHead:
		layerKey = "HD"
		layerValue = object.HD
	case d2enum.CompositeTypeTorso:
		layerKey = "TR"
		layerValue = object.TR
	case d2enum.CompositeTypeLegs:
		layerKey = "LG"
		layerValue = object.LG
	case d2enum.CompositeTypeRightArm:
		layerKey = "RA"
		layerValue = object.RA
	case d2enum.CompositeTypeLeftArm:
		layerKey = "LA"
		layerValue = object.LA
	case d2enum.CompositeTypeRightHand:
		layerKey = "RH"
		layerValue = object.RH
	case d2enum.CompositeTypeLeftHand:
		layerKey = "LH"
		layerValue = object.LH
	case d2enum.CompositeTypeShield:
		layerKey = "SH"
		layerValue = object.SH
	case d2enum.CompositeTypeSpecial1:
		layerKey = "S1"
		layerValue = object.S1
	case d2enum.CompositeTypeSpecial2:
		layerKey = "S2"
		layerValue = object.S2
	case d2enum.CompositeTypeSpecial3:
		layerKey = "S3"
		layerValue = object.S3
	case d2enum.CompositeTypeSpecial4:
		layerKey = "S4"
		layerValue = object.S4
	case d2enum.CompositeTypeSpecial5:
		layerKey = "S5"
		layerValue = object.S5
	case d2enum.CompositeTypeSpecial6:
		layerKey = "S6"
		layerValue = object.S6
	case d2enum.CompositeTypeSpecial7:
		layerKey = "S7"
		layerValue = object.S7
	case d2enum.CompositeTypeSpecial8:
		layerKey = "S8"
		layerValue = object.S8
	default:
		return "", "", errors.New("unknown layer type")
	}

	return layerKey, layerValue, nil
}

func loadCompositeLayer(object *d2datadict.ObjectLookupRecord, layerKey, layerValue, animationMode, weaponClass, palettePath string, transparency int) (*Animation, error) {
	animationPaths := []string{
		fmt.Sprintf("%s/%s/%s/%s%s%s%s%s.dcc", object.Base, object.Token, layerKey, object.Token, layerKey, layerValue, animationMode, weaponClass),
		fmt.Sprintf("%s/%s/%s/%s%s%s%s%s.dcc", object.Base, object.Token, layerKey, object.Token, layerKey, layerValue, animationMode, "HTH"),
		fmt.Sprintf("%s/%s/%s/%s%s%s%s%s.dc6", object.Base, object.Token, layerKey, object.Token, layerKey, layerValue, animationMode, weaponClass),
		fmt.Sprintf("%s/%s/%s/%s%s%s%s%s.dc6", object.Base, object.Token, layerKey, object.Token, layerKey, layerValue, animationMode, "HTH"),
	}

	for _, animationPath := range animationPaths {
		if exists, _ := FileExists(animationPath); exists {
			animation, err := LoadAnimationWithTransparency(animationPath, palettePath, transparency)
			if err == nil {
				return animation, nil
			}
		}
	}

	return nil, errors.New("animation not found")
}
