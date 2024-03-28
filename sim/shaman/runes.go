package shaman

import (
	"slices"
	"time"

	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/proto"
	"github.com/wowsims/sod/sim/core/stats"
)

func (shaman *Shaman) ApplyRunes() {
	// Chest
	shaman.applyDualWieldSpec()
	// shaman.applyHealingRain()
	// shaman.applyOverload()
	shaman.applyShieldMastery()
	shaman.applyTwoHandedMastery()

	// Hands
	shaman.applyLavaBurst()
	shaman.applyLavaLash()
	shaman.applyMoltenBlast()

	// Waist
	shaman.applyFireNova()
	shaman.applyMaelstromWeapon()
	shaman.applyPowerSurge()

	// Legs
	shaman.applyAncestralGuidance()
	// shaman.applyEarthShield()
	shaman.applyShamanisticRage()
	shaman.applyWayOfEarth()

	// Feet
	shaman.applyAncestralAwakening()
	// shaman.applyDecoyTotem()
	shaman.applySpiritOfTheAlpha()
}

func (shaman *Shaman) applyDualWieldSpec() {
	if !shaman.HasRune(proto.ShamanRune_RuneChestDualWieldSpec) || !shaman.HasMHWeapon() || !shaman.HasOHWeapon() {
		return
	}

	shaman.AutoAttacks.OHConfig().DamageMultiplier *= 1.5

	meleeHit := float64(core.MeleeHitRatingPerHitChance * 10)
	spellHit := float64(core.SpellHitRatingPerHitChance * 10)

	shaman.AddStat(stats.MeleeHit, meleeHit)
	shaman.AddStat(stats.SpellHit, spellHit)

	dwBonusApplied := true

	shaman.RegisterAura(core.Aura{
		Label:    "DW Spec Trigger",
		ActionID: core.ActionID{SpellID: int32(proto.ShamanRune_RuneChestDualWieldSpec)},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		// Perform additional checks for later weapon-swapping
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeMH) {
				return
			}

			if shaman.HasMHWeapon() && shaman.HasOHWeapon() {
				if dwBonusApplied {
					return
				} else {
					shaman.AddStat(stats.MeleeHit, meleeHit)
					shaman.AddStat(stats.SpellHit, spellHit)
				}
			} else {
				shaman.AddStat(stats.MeleeHit, -1*meleeHit)
				shaman.AddStat(stats.SpellHit, -1*spellHit)
				dwBonusApplied = false
			}
		},
	})
}

// TODO: Not functional
func (shaman *Shaman) applyShieldMastery() {
	if !shaman.HasRune(proto.ShamanRune_RuneChestShieldMastery) {
		return
	}

	shaman.RegisterAura(core.Aura{
		Label:    "Shield Mastery",
		ActionID: core.ActionID{SpellID: int32(proto.ShamanRune_RuneChestShieldMastery)},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
	})
}

func (shaman *Shaman) applyTwoHandedMastery() {
	if !shaman.HasRune(proto.ShamanRune_RuneTwoHandedMastery) {
		return
	}

	procSpellId := int32(436365)

	// Two-handed mastery gives +10% AP, +30% attack speed, and +10% spell hit
	attackSpeedMultiplier := 1.3
	apMultiplier := 1.1
	spellHitIncrease := core.SpellHitRatingPerHitChance * 10.0

	statDep := shaman.NewDynamicMultiplyStat(stats.AttackPower, apMultiplier)
	procAura := shaman.RegisterAura(core.Aura{
		Label:    "Two-Handed Mastery Proc",
		ActionID: core.ActionID{SpellID: procSpellId},
		Duration: time.Second * 10,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyMeleeSpeed(sim, attackSpeedMultiplier)
			shaman.AddStatDynamic(sim, stats.SpellHit, spellHitIncrease)
			aura.Unit.EnableDynamicStatDep(sim, statDep)
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			shaman.MultiplyAttackSpeed(sim, 1/attackSpeedMultiplier)
			shaman.AddStatDynamic(sim, stats.SpellHit, -1*spellHitIncrease)
			aura.Unit.DisableDynamicStatDep(sim, statDep)
		},
	})

	shaman.RegisterAura(core.Aura{
		Label:    "Two-Handed Mastery",
		ActionID: core.ActionID{SpellID: int32(proto.ShamanRune_RuneTwoHandedMastery)},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || !spell.ProcMask.Matches(core.ProcMaskMeleeMH) {
				return
			}

			if shaman.MainHand().HandType == proto.HandType_HandTypeTwoHand {
				procAura.Activate(sim)
			} else {
				procAura.Deactivate(sim)
			}
		},
	})
}

func (shaman *Shaman) applyMaelstromWeapon() {
	if !shaman.HasRune(proto.ShamanRune_RuneWaistMaelstromWeapon) {
		return
	}

	buffSpellId := 408505
	buffDuration := time.Second * 30

	ppm := core.TernaryFloat64(shaman.GetCharacter().Consumes.MainHandImbue == proto.WeaponImbue_WindfuryWeapon, 15, 10)

	var affectedSpells []*core.Spell
	var affectedSpellCodes = []int32{
		SpellCode_ShamanLightningBolt,
		SpellCode_ShamanChainLightning,
		SpellCode_ShamanLavaBurst,
		SpellCode_ShamanHealingWave,
		SpellCode_ShamanLesserHealingWave,
		SpellCode_ShamanChainHeal,
	}

	shaman.MaelstromWeaponAura = shaman.RegisterAura(core.Aura{
		Label:     "MaelstromWeapon Proc",
		ActionID:  core.ActionID{SpellID: int32(buffSpellId)},
		Duration:  buffDuration,
		MaxStacks: 5,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(
				core.Flatten([][]*core.Spell{
					shaman.LightningBolt,
					shaman.ChainLightning,
					{shaman.LavaBurst},
					shaman.HealingWave,
					shaman.LesserHealingWave,
					shaman.ChainHeal,
				}), func(spell *core.Spell) bool { return spell != nil },
			)
		},
		OnStacksChange: func(aura *core.Aura, sim *core.Simulation, oldStacks int32, newStacks int32) {
			multDiff := 0.2 * float64(newStacks-oldStacks)
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier -= multDiff })
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CostMultiplier -= multDiff })
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !slices.Contains(affectedSpellCodes, spell.SpellCode) {
				return
			}

			shaman.MaelstromWeaponAura.Deactivate(sim)
		},
	})

	ppmm := shaman.AutoAttacks.NewPPMManager(ppm, core.ProcMaskMelee)

	// This aura is hidden, just applies stacks of the proc aura.
	shaman.RegisterAura(core.Aura{
		Label:    "MaelstromWeapon",
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() {
				return
			}

			if ppmm.Proc(sim, spell.ProcMask, "Maelstrom Weapon") {
				shaman.MaelstromWeaponAura.Activate(sim)
				shaman.MaelstromWeaponAura.AddStack(sim)
			}
		},
	})
}

const ShamanPowerSurgeProcChance = .05

func (shaman *Shaman) applyPowerSurge() {
	if !shaman.HasRune(proto.ShamanRune_RuneWaistPowerSurge) {
		return
	}

	// TODO: Figure out how this actually works becaue the 2024-02-27 tuning notes make it sound like
	// this is not just a fully passive stat boost
	shaman.AddStat(stats.MP5, shaman.GetStat(stats.Intellect)*.15)

	var affectedSpells []*core.Spell
	var affectedSpellCodes = []int32{
		SpellCode_ShamanChainLightning,
		SpellCode_ShamanChainHeal,
		SpellCode_ShamanLavaBurst,
	}

	shaman.PowerSurgeAura = shaman.RegisterAura(core.Aura{
		Label:    "Power Surge Proc",
		ActionID: core.ActionID{SpellID: 440285},
		Duration: time.Second * 10,
		OnInit: func(aura *core.Aura, sim *core.Simulation) {
			affectedSpells = core.FilterSlice(
				core.Flatten([][]*core.Spell{
					shaman.ChainLightning,
					shaman.ChainHeal,
					{shaman.LavaBurst},
				}), func(spell *core.Spell) bool { return spell != nil })
		},
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) {
				spell.CastTimeMultiplier -= 1
				if spell.CD.Timer != nil {
					spell.CD.Reset()
				}
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			core.Each(affectedSpells, func(spell *core.Spell) { spell.CastTimeMultiplier += 1 })
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !slices.Contains(affectedSpellCodes, spell.SpellCode) {
				return
			}
			aura.Deactivate(sim)
		},
	})
}

func (shaman *Shaman) applyWayOfEarth() {
	if !shaman.HasRune(proto.ShamanRune_RuneLegsWayOfEarth) {
		return
	}

	if shaman.Consumes.MainHandImbue == proto.WeaponImbue_RockbiterWeapon {
		shaman.PseudoStats.ThreatMultiplier *= 1.5
	}

	shaman.RegisterAura(core.Aura{
		Label:    "Way of Earth",
		ActionID: core.ActionID{SpellID: int32(proto.ShamanRune_RuneLegsWayOfEarth)},
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			aura.Activate(sim)
		},
	})
}

func (shaman *Shaman) applySpiritOfTheAlpha() {
	if !shaman.HasRune(proto.ShamanRune_RuneFeetSpiritOfTheAlpha) {
		return
	}

	// Spirit of the Alpha currently gives +5% physical damage when used on another target.
	// Assume this as a default
	shaman.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= 1.05
}
