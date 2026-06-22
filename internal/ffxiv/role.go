package ffxiv

import (
	"reflect"
)

type Role int

const (
	DPS Role = iota
	Healer
	Tank
	Empty
)

type Roles struct {
	Roles []Role
}

func (rs Roles) Emoji() string {
	if reflect.DeepEqual(rs.Roles, []Role{DPS}) {
		return "<:dps:1518605449962848276>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{Healer}) {
		return "<:healer:1518605448998293715>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{Tank}) {
		return "<:tank:1518605472733855835>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{DPS, Healer}) {
		return "<:healerdps:1518614268453585067>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{DPS, Tank}) {
		return "<:tankdps:1518614269397303356>"
	}
	if reflect.DeepEqual(rs.Roles, []Role{Healer, Tank}) {
		return "<:tankhealer:1518614266045923459>"
	}

	if reflect.DeepEqual(rs.Roles, []Role{Healer, Tank, DPS}) {
		return "<:tankhealerdps:1518614267375390770>"
	}

	return "<:tankhealerdps:1518614267375390770>"
}
