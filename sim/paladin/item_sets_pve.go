package paladin

import (
	"time"

	"github.com/wowsims/sod/sim/core"
	"github.com/wowsims/sod/sim/core/stats"
)

///////////////////////////////////////////////////////////////////////////
//                            SoD Phase 3 Item Sets
///////////////////////////////////////////////////////////////////////////

var ItemSetObsessedProphetsPlate = core.NewItemSet(core.ItemSet{
	Name: "Obsessed Prophet's Plate",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.MeleeCrit, 1*core.CritRatingPerCritChance)
			c.AddStat(stats.SpellCrit, 1*core.SpellCritRatingPerCritChance)
		},
		3: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.PseudoStats.SchoolBonusCritChance[stats.SchoolIndexHoly] += 3 * core.SpellCritRatingPerCritChance
		},
	},
})

var _ = core.NewItemSet(core.ItemSet{
	Name: "Emerald Encrusted Battleplate",
	Bonuses: map[int32]core.ApplyEffect{
		3: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Stamina, 10)
		},
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.HealingPower, 22)
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            SoD Phase 4 Item Sets
///////////////////////////////////////////////////////////////////////////

var ItemSetSoulforgeArmor = core.NewItemSet(core.ItemSet{
	Name: "Soulforge Armor",
	Bonuses: map[int32]core.ApplyEffect{
		// +40 Attack Power and up to 40 increased healing from spells.
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.AttackPower:       40,
				stats.RangedAttackPower: 40,
				stats.HealingPower:      40,
			})
		},
		// 6% chance on melee autoattack and 4% chance on spellcast to increase your damage and healing done by magical spells and effects by up to 95 for 10 sec.
		4: func(agent core.Agent) {
			c := agent.GetCharacter()
			actionID := core.ActionID{SpellID: 450625}

			procAura := c.NewTemporaryStatsAura("Crusader's Wrath", core.ActionID{SpellID: 27499}, stats.Stats{stats.SpellPower: 95}, time.Second*10)
			handler := func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
				procAura.Activate(sim)
			}

			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID:   actionID,
				Name:       "Item - Crusader's Wrath Proc - Lightforge Armor (Melee Auto)",
				Callback:   core.CallbackOnSpellHitDealt,
				Outcome:    core.OutcomeLanded,
				ProcMask:   core.ProcMaskMeleeWhiteHit,
				ProcChance: 0.06,
				Handler:    handler,
			})
			core.MakeProcTriggerAura(&c.Unit, core.ProcTrigger{
				ActionID:   actionID,
				Name:       "Item - Crusader's Wrath Proc - Lightforge Armor (Spell Cast)",
				Callback:   core.CallbackOnCastComplete,
				ProcMask:   core.ProcMaskSpellDamage | core.ProcMaskSpellHealing,
				ProcChance: 0.04,
				Handler:    handler,
			})
		},
		// +8 All Resistances.
		6: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddResistances(8)
		},
		// +200 Armor.
		8: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStat(stats.Armor, 200)
		},
	},
})

var ItemSetLawbringerRadiance = core.NewItemSet(core.ItemSet{
	Name: "Lawbringer Radiance",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// No need to model
			//(2) Set : Your Judgement of Light and Judgement of Wisdom also grant the effects of Judgement of the Crusader.
		},
		4: func(agent core.Agent) {
			character := agent.GetCharacter()
			character.AddStat(stats.MeleeCrit, 2)
			character.AddStat(stats.SpellCrit, 2)
		},
		6: func(agent core.Agent) {
			// Implemented in Paladin.go
			paladin := agent.(PaladinAgent).GetPaladin()
			core.MakePermanent(paladin.RegisterAura(core.Aura{
				Label: "S03 - Item - T1 - Paladin - Retribution 6P Bonus",
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					paladin.lingerDuration = time.Second * 6
				},
			}))
		},
	},
})

///////////////////////////////////////////////////////////////////////////
//                            SoD Phase 5 Item Sets
///////////////////////////////////////////////////////////////////////////

var ItemSetFreethinkersArmor = core.NewItemSet(core.ItemSet{
	Name: "Freethinker's Armor",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			c := agent.GetCharacter()
			c.AddStats(stats.Stats{
				stats.HolyPower: 14,
			})
		},
		3: func(agent core.Agent) {
			// Increases damage done by your holy shock spell by 50%
			paladin := agent.GetCharacter()
			paladin.OnSpellRegistered(func(spell *core.Spell) {
				if spell.SpellCode == SpellCode_PaladinHolyShock {
					spell.DamageMultiplier *= 1.5
				}
			})
		},
		5: func(agent core.Agent) {
			// Reduce cooldown of Exorcism by 3 seconds
			paladin := agent.(PaladinAgent).GetPaladin()
			paladin.RegisterAura(core.Aura{
				Label: "S03 - Item - ZG - Paladin - Caster 5P Bonus",
				OnInit: func(aura *core.Aura, sim *core.Simulation) {
					for _, spell := range paladin.exorcism {
						spell.CD.Duration -= time.Second * 3
					}
				},
			})
		},
	},
})

var ItemSetMercifulJudgement = core.NewItemSet(core.ItemSet{
	Name: "Merciful Judgement",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			//Increases critical strike chance of holy shock spell by 20%
			paladin := agent.GetCharacter()
			paladin.OnSpellRegistered(func(spell *core.Spell) {
				if spell.SpellCode == SpellCode_PaladinHolyShock {
					spell.BonusCritRating += 20.0
				}
			})
		},
		4: func(agent core.Agent) {
			//Increases damage done by your Consecration spell by 50%
			paladin := agent.GetCharacter()
			paladin.OnSpellRegistered(func(spell *core.Spell) {
				if spell.SpellCode == SpellCode_PaladinConsecration {
					spell.DamageMultiplier *= 1.5
				}
			})
		},
		6: func(agent core.Agent) {
			// While you are not your Beacon of Light target, your Beacon of Light target is also healed by 100% of the damage you deal
			// with Consecration, Exorcism, Holy Shock, Holy Wrath, and Hammer of Wrath
			// No need to Sim
		},
	},
})

var ItemSetRadiantJudgement = core.NewItemSet(core.ItemSet{
	Name: "Radiant Judgement",
	Bonuses: map[int32]core.ApplyEffect{
		2: func(agent core.Agent) {
			// 2 pieces: Increases damage done by your damaging Judgements by 20% and your Judgements no longer consume your Seals on the target.
			paladin := agent.(PaladinAgent).GetPaladin()
			core.MakePermanent(paladin.RegisterAura(core.Aura{
				Label: "S03 - Item - T2 - Paladin - Retribution 2P Bonus",
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					if !paladin.t2Judgement2pc {
						paladin.t2Judgement2pc = true
						paladin.enableT2Judgement2pc()
					}
				},
			}))
		},
		4: func(agent core.Agent) {
			// 4 pieces: The cooldown on your Judgement is instantly reset if used on a different Seal than your last Judgement.
			// Implemented in Paladin.go
			paladin := agent.(PaladinAgent).GetPaladin()
			core.MakePermanent(paladin.RegisterAura(core.Aura{
				Label: "S03 - Item - T2 - Paladin - Retribution 4P Bonus",
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					paladin.t2Judgement4pc = true
				},
			}))
		},
		6: func(agent core.Agent) {
			// 6 pieces: Your Judgement grants 1% increased Holy damage for 8 sec, stacking up to 5 times.
			// Implemented in Paladin.go
			paladin := agent.(PaladinAgent).GetPaladin()
			core.MakePermanent(paladin.RegisterAura(core.Aura{
				Label: "S03 - Item - T2 - Paladin - Retribution 6P Bonus",
				OnReset: func(aura *core.Aura, sim *core.Simulation) {
					if !paladin.t2Judgement6pc {
						paladin.t2Judgement6pc = true
						paladin.enableT2Judgement6pc()
					}
				},
			}))
		},
	},
})
