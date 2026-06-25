import type { ListingEntry } from '#lib/scraper/types';

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

export enum FindingwayEvents {
	PostMerc = 'postMerc',
	PostTea = 'postTea',
	PostUcob = 'postUcob',
	PostUwu = 'postUwu'
}

export const enum BrandingColors {
	Primary = 0x6684927
}

export const enum LanguageFormatters {
	Date = 'date',
	InlineCode = 'inlineCode',
	Number = 'number',
	Permissions = 'permissions',
	RelativeTime = 'relativeTime',
	Time = 'time'
}

export const enum ErrorIdentifiers {
	SetNoChannelConfigured = 'SetNoChannelConfigured'
}

export interface PostMessagePayload {
	entries: ListingEntry[];
	timeOfScrape: number;
}

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

export type Emojis = Classes | Jobs | Roles | 'GreenTick' | 'RedCross';
