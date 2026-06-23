export const rootFolder = new URL('../../../', import.meta.url);

/**
 * Left-to-right mark character.
 *
 * @see {@link https://en.wikipedia.org/wiki/Left-to-right_mark}
 */
export const leftToRightMark = String.fromCodePoint(8_206);

/**
 * Braille pattern blank character
 *
 * @see {@link https://en.wikipedia.org/wiki/Braille_Patterns}
 */
export const braillePatternBlank = String.fromCodePoint(10_240);

/**
 * Zero Width Space character
 *
 * @see {@link https://en.wikipedia.org/wiki/Zero-width_space}
 */
export const zeroWidthSpace = String.fromCodePoint(8_203);

export const enum XIVServers {
	Aether = 'aether',
	Chaos = 'chaos',
	Crystal = 'crystal',
	Dynamis = 'dynamis',
	Elemental = 'elemental',
	Gaia = 'gaia',
	Light = 'light',
	Mana = 'mana',
	Materia = 'materia',
	Meteor = 'Meteor',
	Primal = 'primal'
}

export const enum BrandingColors {
	Primary = 0xbb77ea,
	ExpiredEvent = 0xff0000
}

export const enum ErrorIdentifiers {}

/* eslint-disable typescript-sort-keys/string-enum */
export const enum CustomIdPrefixes {}
/* eslint-enable typescript-sort-keys/string-enum */

export type Classes = 'Arcanist' | 'Archer' | 'Conjurer' | 'Gladiator' | 'Lancer' | 'Marauder' | 'Pugilist' | 'Rogue' | 'Thaumaturge';

export type Roles = 'DPS' | 'Healer' | 'HealerDPS' | 'Tank' | 'TankDPS' | 'TankHealer' | 'TankHealerDps';

export type Jobs =
	| 'Astrologian'
	| 'Bard'
	| 'BlackMage'
	| 'BlueMage'
	| 'Dancer'
	| 'DarkKnight'
	| 'Dragoon'
	| 'Gunbreaker'
	| 'Machinist'
	| 'Monk'
	| 'Ninja'
	| 'Paladin'
	| 'Pictomancer'
	| 'Reaper'
	| 'RedMage'
	| 'Sage'
	| 'Samurai'
	| 'Scholar'
	| 'Summoner'
	| 'Viper'
	| 'Warrior'
	| 'WhiteMage';

export type Emojis = Classes | Roles | Jobs;
