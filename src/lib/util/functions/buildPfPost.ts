import { ChannelType } from '#lib/generated/prisma-client/enums';
import type { Party } from '#lib/scraper/types';
import type { PostMessagePayload } from '#lib/util/constants';
import { FindingwayEmojis } from '#utils/emojis';
import { resolveKey } from '@sapphire/plugin-i18next';
import { roundNumber } from '@sapphire/utilities';
import { bold, codeBlock, ContainerBuilder, time, TimestampStyles } from 'discord.js';

export async function buildPfPost({ entries, timeOfScrape }: PostMessagePayload, type: ChannelType): Promise<ContainerBuilder> {
	const container = new ContainerBuilder();
	container.setAccentColor(getAccentColour(type));

	const title = await resolveKey(null!, `pfposts:title.${type}`, { lng: 'en-US' });
	const pfs = await resolveKey(null!, `pfposts:pfs`, { lng: 'en-US' });

	container.addTextDisplayComponents(
		(textDisplay) => textDisplay.setContent(bold(`${title} ${pfs} [Light]`)),
		(textDisplay) =>
			textDisplay.setContent(
				`Found ${bold(entries.length.toString())} active listings • ${time(roundNumber(timeOfScrape / 1_000), TimestampStyles.RelativeTime)}`
			)
	);

	for (const [index, entry] of entries.entries()) {
		const minIlvl = await resolveKey(null!, `pfposts:minilvl`, { lng: 'en-US', minilvl: entry.minIlvl });

		const row = container.addTextDisplayComponents(
			(textDisplay) =>
				textDisplay.setContent(
					[
						//
						bold(entry.creator),
						bold(minIlvl),
						bold(entry.pfTags ?? '')
					].join('\t')
				),
			(textDisplay) =>
				textDisplay.setContent(
					[
						//
						buildPartyListing(entry.party),
						bold(`⌛ ${time(entry.expires, TimestampStyles.RelativeTime)}`),
						bold(`⏱️ ${time(entry.updated, TimestampStyles.RelativeTime)}`)
					].join('\t')
				),
			(textDisplay) => textDisplay.setContent(codeBlock('txt', entry.description))
		);

		if (index < entries.length - 1) {
			row.addSeparatorComponents((separator) => separator);
		}
	}

	return container;
}

function buildPartyListing(party: Party[]): string {
	return party
		.map((slot) => {
			if (slot.filled && slot.title) {
				return partyTitlesToEmojis(slot.title);
			}

			return partyTypesToEmojis(slot.type);
		})
		.join('');
}

function partyTitlesToEmojis(title: string): string {
	switch (title) {
		// Jobs
		case 'AST':
			return FindingwayEmojis.Astrologian;
		case 'BLM':
			return FindingwayEmojis.BlackMage;
		case 'BLU':
			return FindingwayEmojis.BlueMage;
		case 'BRD':
			return FindingwayEmojis.Bard;
		case 'DNC':
			return FindingwayEmojis.Dancer;
		case 'DRG':
			return FindingwayEmojis.Dragoon;
		case 'DRK':
			return FindingwayEmojis.DarkKnight;
		case 'GNB':
			return FindingwayEmojis.Gunbreaker;
		case 'MCH':
			return FindingwayEmojis.Machinist;
		case 'MNK':
			return FindingwayEmojis.Monk;
		case 'NIN':
			return FindingwayEmojis.Ninja;
		case 'PCT':
			return FindingwayEmojis.Pictomancer;
		case 'PLD':
			return FindingwayEmojis.Paladin;
		case 'RDM':
			return FindingwayEmojis.RedMage;
		case 'RPR':
			return FindingwayEmojis.Reaper;
		case 'SAM':
			return FindingwayEmojis.Samurai;
		case 'SCH':
			return FindingwayEmojis.Scholar;
		case 'SGE':
			return FindingwayEmojis.Sage;
		case 'SMN':
			return FindingwayEmojis.Summoner;
		case 'VPR':
			return FindingwayEmojis.Viper;
		case 'WAR':
			return FindingwayEmojis.Warrior;
		case 'WHM':
			return FindingwayEmojis.WhiteMage;

		// Classes
		case 'ACN':
			return FindingwayEmojis.Arcanist;
		case 'ARC':
			return FindingwayEmojis.Archer;
		case 'CNJ':
			return FindingwayEmojis.Conjurer;
		case 'GLD':
			return FindingwayEmojis.Gladiator;
		case 'LNC':
			return FindingwayEmojis.Lancer;
		case 'MRD':
			return FindingwayEmojis.Marauder;
		case 'PUG':
			return FindingwayEmojis.Pugilist;
		case 'ROG':
			return FindingwayEmojis.Rogue;
		case 'THM':
			return FindingwayEmojis.Thaumaturge;
		default:
			return FindingwayEmojis.TankHealerDps;
	}
}

function partyTypesToEmojis(type: string): string {
	switch (type) {
		case 'dps':
			return FindingwayEmojis.DPS;
		case 'healer':
			return FindingwayEmojis.Healer;
		case 'healerDps':
			return FindingwayEmojis.HealerDPS;
		case 'tank':
			return FindingwayEmojis.Tank;
		case 'tankDps':
			return FindingwayEmojis.TankDPS;
		case 'tankHealer':
			return FindingwayEmojis.TankHealer;
		case 'tankHealerDps':
			return FindingwayEmojis.TankHealerDps;
		default:
			return '';
	}
}

function getAccentColour(type: ChannelType): number {
	switch (type) {
		case ChannelType.TheEpicOfAlexander:
			return 0xffd833;
		case ChannelType.TheWeaponsRefrain:
			return 0x5a9bff;
		case ChannelType.TheUnendingCoilOfBahamut:
			return 0xffa752;
		case ChannelType.Mercantile:
			return 0xbe20cd;
	}
}
