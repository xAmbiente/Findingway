import type { Emojis, Jobs } from '#lib/util/constants';
import { formatEmoji } from 'discord.js';

export const FindingwayEmojis: Record<Emojis, string> = {
	// Jobs
	Astrologian: formatEmoji('1518605444044816447', false),
	Bard: formatEmoji('1518605456807821342', false),
	BlackMage: formatEmoji('1518605442849181768', false),
	BlueMage: formatEmoji('1518605465444155412', false),
	Dancer: formatEmoji('1518605471194288230', false),
	DarkKnight: formatEmoji('1518605466509250701', false),
	Dragoon: formatEmoji('1518605468702871563', false),
	Gunbreaker: formatEmoji('1518605453284741222', false),
	Machinist: formatEmoji('1518605467608416396', false),
	Monk: formatEmoji('1518605454748549120', false),
	Ninja: formatEmoji('1518605457755869316', false),
	Paladin: formatEmoji('1518605461891579977', false),
	Pictomancer: formatEmoji('1518605446666129478', false),
	Reaper: formatEmoji('1518605464009576490', false),
	RedMage: formatEmoji('1518605459806748774', false),
	Sage: formatEmoji('1518605451351036036', false),
	Samurai: formatEmoji('1518605445395382435', false),
	Scholar: formatEmoji('1518605460729626858', false),
	Summoner: formatEmoji('1518605469827072165', false),
	Viper: formatEmoji('1518605463250407434', false),
	Warrior: formatEmoji('1518605447970426880', false),
	WhiteMage: formatEmoji('1518605441905594428', false),

	// Classes
	Arcanist: formatEmoji('1518612053315682509', false),
	Archer: formatEmoji('1518612061217882222', false),
	Conjurer: formatEmoji('1518612064594166016', false),
	Gladiator: formatEmoji('1518612056444764375', false),
	Lancer: formatEmoji('1518612058067959908', false),
	Marauder: formatEmoji('1518612059762200777', false),
	Pugilist: formatEmoji('1518612062773969068', false),
	Rogue: formatEmoji('1518612051646353498', false),
	Thaumaturge: formatEmoji('1518612055232479254', false),

	// Roles
	DPS: formatEmoji('1518605449962848276', false),
	Healer: formatEmoji('1518605448998293715', false),
	HealerDPS: formatEmoji('1518614268453585067', false),
	Tank: formatEmoji('1518605472733855835', false),
	TankDPS: formatEmoji('1518614269397303356', false),
	TankHealer: formatEmoji('1518614266045923459', false),
	TankHealerDps: formatEmoji('1518614267375390770', false),

	// Other
	RedCross: formatEmoji({ id: '1518970626797473984', name: 'redcross' }),
	GreenTick: formatEmoji({ id: '1518970625216221224', name: 'greentick' })
} as const;

export function getEmojiForJob(job: Jobs): string {
	return FindingwayEmojis[job] || '';
}
