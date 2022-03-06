package rpcclient

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type MechContainer struct {
	Mech    *Mech
	Chassis *Chassis
	Weapons map[int]*Weapon
	Turrets map[int]*Weapon
	Modules map[int]*Module
}

type Mech struct {
	ID           string
	OwnerID      string
	TemplateID   string
	ChassisID    string
	Tier         string
	IsDefault    bool
	ImageURL     string
	AnimationURL string
	Hash         string
	Name         string
	Label        string
	Slug         string
	DeletedAt    null.Time
	UpdatedAt    time.Time
	CreatedAt    time.Time
}
type Chassis struct {
	ID                 string
	BrandID            string
	Label              string
	Model              string
	Skin               string
	Slug               string
	ShieldRechargeRate int
	HealthRemaining    int
	WeaponHardpoints   int
	TurretHardpoints   int
	UtilitySlots       int
	Speed              int
	MaxHitpoints       int
	MaxShield          int
	DeletedAt          null.Time
	UpdatedAt          time.Time
	CreatedAt          time.Time
}
type Weapon struct {
	ID         string
	BrandID    null.String
	Label      string
	Slug       string
	Damage     int
	WeaponType string
	DeletedAt  null.Time
	UpdatedAt  time.Time
	CreatedAt  time.Time
}
type Module struct {
	ID               string
	BrandID          null.String
	Slug             string
	Label            string
	HitpointModifier int
	ShieldModifier   int
	DeletedAt        null.Time
	UpdatedAt        time.Time
	CreatedAt        time.Time
}
