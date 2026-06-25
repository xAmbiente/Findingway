import { ChannelType } from '#lib/generated/prisma-client/enums';

export const SearchTerms: Record<Exclude<ChannelType, 'Mercantile'>, string> = {
	[ChannelType.TheEpicOfAlexander]: 'The Epic of Alexander',
	[ChannelType.TheUnendingCoilOfBahamut]: 'The Unending Coil of Bahamut',
	[ChannelType.TheWeaponsRefrain]: "The Weapon's Refrain"
} as const;
