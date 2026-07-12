import { Owners } from '#root/config';
import { BrandingColors } from '#utils/constants';
import {
	BOOST_PERK_PAYOUT_COORDINATOR_ID,
	BOOST_PERK_REDEMPTION_CHANNEL_ID,
	BOOST_PERK_WEEK_MILLISECONDS,
	BOOST_PERK_WEEKLY_GIL
} from '#utils/constants/boost-perk';
import { container, type ChatInputCommand } from '@sapphire/framework';
import { resolveKey } from '@sapphire/plugin-i18next';
import { ChannelType, ContainerBuilder, MessageFlags, userMention, type ButtonInteraction } from 'discord.js';

export type BoostPerkInteraction = ButtonInteraction<'cached'> | ChatInputCommand.Interaction<'cached'>;

export async function getBoostPerkRedeemStatus(interaction: BoostPerkInteraction): Promise<RedeemStatus> {
	const now = new Date();

	const premiumSinceTimestamp = interaction.member.premiumSinceTimestamp;

	if (!premiumSinceTimestamp) {
		return {
			isPremiumMember: false,
			eligible: false,
			nextEligibleAt: null,
			lastRedeemDate: null,
			premiumSince: null,
			weeksPremium: 0,
			redeemableWeeks: 0,
			redeemableAmount: 0
		};
	}

	const premiumSince = new Date(premiumSinceTimestamp);
	const guildMemberRecord = await container.prisma.guildMember.findUnique({
		where: {
			guildId_userId: {
				guildId: interaction.guildId,
				userId: interaction.user.id
			}
		}
	});

	const lastRedeemDate = guildMemberRecord?.boostRedeemDate ?? null;
	const redeemFromDate = lastRedeemDate && lastRedeemDate > premiumSince ? lastRedeemDate : premiumSince;
	const weeksPremium = calculateFullWeeks(premiumSince, now);
	const redeemableWeeks = calculateFullWeeks(redeemFromDate, now);
	const eligible = redeemableWeeks >= 1;

	return {
		isPremiumMember: true,
		eligible,
		nextEligibleAt: eligible ? null : new Date(redeemFromDate.getTime() + BOOST_PERK_WEEK_MILLISECONDS),
		lastRedeemDate,
		premiumSince,
		weeksPremium,
		redeemableWeeks,
		redeemableAmount: redeemableWeeks * BOOST_PERK_WEEKLY_GIL
	};
}

export async function runBoostPerkRedeemFlow(interaction: BoostPerkInteraction) {
	const status = await getBoostPerkRedeemStatus(interaction);

	if (!status.eligible) {
		if (!status.isPremiumMember) {
			const content = await resolveKey(interaction, 'commands/boost-perk:redeemNotPremium');
			const responseContainer = new ContainerBuilder()
				.setAccentColor(BrandingColors.Primary)
				.addTextDisplayComponents((textDisplay) => textDisplay.setContent(content));

			return interaction.editReply({
				components: [responseContainer],
				flags: MessageFlags.IsComponentsV2
			});
		}

		const content = await resolveKey(interaction, 'commands/boost-perk:redeemCooldown', {
			nextEligibleAt: status.nextEligibleAt
		});
		const responseContainer = new ContainerBuilder()
			.setAccentColor(BrandingColors.Primary)
			.addTextDisplayComponents((textDisplay) => textDisplay.setContent(content));

		return interaction.editReply({
			components: [responseContainer],
			flags: MessageFlags.IsComponentsV2
		});
	}

	const channel = await interaction.guild.channels.fetch(BOOST_PERK_REDEMPTION_CHANNEL_ID);

	if (channel?.type !== ChannelType.GuildText) {
		const content = await resolveKey(interaction, 'commands/boost-perk:redeemThreadCreateFailed');
		const responseContainer = new ContainerBuilder()
			.setAccentColor(BrandingColors.Primary)
			.addTextDisplayComponents((textDisplay) => textDisplay.setContent(content));

		return interaction.editReply({
			components: [responseContainer],
			flags: MessageFlags.IsComponentsV2
		});
	}

	const now = new Date();
	const thread = await channel.threads.create({
		name: await resolveKey(interaction, 'commands/boost-perk:redeemThreadName', {
			userId: interaction.user.id,
			epoch: Math.floor(now.getTime() / 1_000)
		}),
		type: ChannelType.PrivateThread,
		invitable: false,
		reason: await resolveKey(interaction, 'commands/boost-perk:redeemThreadReason', {
			userTag: interaction.user.tag
		})
	});

	const memberIds = new Set([interaction.user.id, ...Owners]);
	await Promise.all([...memberIds].map(async (memberId) => thread.members.add(memberId).catch(() => null)));

	const threadSetupMessage = await resolveKey(interaction, 'commands/boost-perk:redeemThreadSetupMessage', {
		redeemerMention: interaction.user.toString(),
		payoutCoordinatorMention: userMention(BOOST_PERK_PAYOUT_COORDINATOR_ID)
	});

	await thread.send({
		content: threadSetupMessage
	});

	await container.prisma.guildMember.upsert({
		where: {
			guildId_userId: {
				guildId: interaction.guildId,
				userId: interaction.user.id
			}
		},
		update: {
			boostRedeemDate: now
		},
		create: {
			guildId: interaction.guildId,
			userId: interaction.user.id,
			boostRedeemDate: now
		}
	});

	const content = await resolveKey(interaction, 'commands/boost-perk:redeemStarted', {
		threadMention: thread.toString(),
		count: status.redeemableWeeks,
		redeemableWeeks: status.redeemableWeeks,
		amount: status.redeemableAmount
	});
	const responseContainer = new ContainerBuilder()
		.setAccentColor(BrandingColors.Primary)
		.addTextDisplayComponents((textDisplay) => textDisplay.setContent(content));

	return interaction.editReply({
		components: [responseContainer],
		flags: MessageFlags.IsComponentsV2
	});
}

function calculateFullWeeks(from: Date, to: Date): number {
	const elapsedMilliseconds = Math.max(0, to.getTime() - from.getTime());
	return Math.floor(elapsedMilliseconds / BOOST_PERK_WEEK_MILLISECONDS);
}

interface RedeemStatus {
	eligible: boolean;
	isPremiumMember: boolean;
	lastRedeemDate: Date | null;
	nextEligibleAt: Date | null;
	premiumSince: Date | null;
	redeemableAmount: number;
	redeemableWeeks: number;
	weeksPremium: number;
}
