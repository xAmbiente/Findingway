import { BrandingColors } from '#utils/constants';
import { BOOST_PERK_WEEKLY_GIL } from '#utils/constants/boost-perk';
import { calculateFullWeeks } from '#utils/functions/boostPerk';
import { ApplyOptions, RegisterChatInputCommand } from '@sapphire/decorators';
import { isGuildMember } from '@sapphire/discord.js-utilities';
import { Command, container, type ChatInputCommand } from '@sapphire/framework';
import { applyLocalizedBuilder, resolveKey } from '@sapphire/plugin-i18next';
import { ApplicationIntegrationType, ContainerBuilder, MessageFlags, PermissionFlagsBits, userMention } from 'discord.js';

@ApplyOptions<ChatInputCommand.Options>({
	requiredClientPermissions: [PermissionFlagsBits.SendMessages]
})
@RegisterChatInputCommand((builder) =>
	applyLocalizedBuilder(builder, 'commands/check-boost-perks:root')
		.setIntegrationTypes(ApplicationIntegrationType.GuildInstall)
		.setDefaultMemberPermissions(PermissionFlagsBits.Administrator)
)
export class SlashCommand extends Command {
	public override async chatInputRun(interaction: ChatInputCommand.Interaction) {
		await interaction.deferReply({ flags: MessageFlags.Ephemeral });

		if (!interaction.inCachedGuild() || !isGuildMember(interaction.member)) {
			return interaction.editReply({
				content: await resolveKey(interaction, 'commands/check-boost-perks:genericMemberError')
			});
		}

		const guildMembers = await interaction.guild.members.fetch();
		const boosters = guildMembers.filter((member) => member.premiumSinceTimestamp !== null);

		if (boosters.size === 0) {
			const emptyMessage = await resolveKey(interaction, 'commands/check-boost-perks:empty');
			const containerMessage = new ContainerBuilder()
				.setAccentColor(BrandingColors.Primary)
				.addTextDisplayComponents((textDisplay) => textDisplay.setContent(emptyMessage));

			return interaction.editReply({
				components: [containerMessage],
				flags: MessageFlags.IsComponentsV2
			});
		}

		const guildMemberRecords = await container.prisma.guildMember.findMany({
			where: {
				guildId: interaction.guildId
			},
			select: {
				userId: true,
				boostRedeemDate: true
			}
		});

		const lastRedeemDateByUserId = new Map<string, Date>(guildMemberRecords.map((record) => [record.userId, record.boostRedeemDate]));
		const now = new Date();
		const boosterRows = boosters
			.map((member) => {
				const premiumSinceTimestamp = member.premiumSinceTimestamp;

				if (premiumSinceTimestamp === null) {
					return null;
				}

				const premiumSince = new Date(premiumSinceTimestamp);
				const lastRedeemDate = lastRedeemDateByUserId.get(member.id) ?? null;
				const redeemFromDate = lastRedeemDate !== null && lastRedeemDate > premiumSince ? lastRedeemDate : premiumSince;
				const boostingWeeks = calculateFullWeeks(premiumSince, now);
				const redeemableWeeks = calculateFullWeeks(redeemFromDate, now);

				return {
					boostingWeeks,
					lastRedeemDate,
					username: member.user.username,
					mention: userMention(member.id),
					premiumSince,
					redeemableAmount: redeemableWeeks * BOOST_PERK_WEEKLY_GIL,
					redeemableWeeks
				};
			})
			.filter((boosterRow) => boosterRow !== null)
			.sort((left, right) => {
				if (left.redeemableAmount !== right.redeemableAmount) {
					return right.redeemableAmount - left.redeemableAmount;
				}

				if (left.redeemableWeeks !== right.redeemableWeeks) {
					return right.redeemableWeeks - left.redeemableWeeks;
				}

				return left.premiumSince.getTime() - right.premiumSince.getTime();
			});

		const title = await resolveKey(interaction, 'commands/check-boost-perks:title');
		const subtitle = await resolveKey(interaction, 'commands/check-boost-perks:subtitle', {
			weeklyGil: BOOST_PERK_WEEKLY_GIL
		});
		const tableHeader = await resolveKey(interaction, 'commands/check-boost-perks:tableHeader');

		const responseContainer = new ContainerBuilder().setAccentColor(BrandingColors.Primary);
		responseContainer.addTextDisplayComponents(
			(textDisplay) => textDisplay.setContent(title),
			(textDisplay) => textDisplay.setContent(subtitle),
			(textDisplay) => textDisplay.setContent(tableHeader)
		);

		for (const row of boosterRows) {
			const lastRedeemed =
				row.lastRedeemDate === null
					? await resolveKey(interaction, 'commands/check-boost-perks:neverRedeemed')
					: await resolveKey(interaction, 'commands/check-boost-perks:lastRedeemedValue', {
							lastRedeemDate: row.lastRedeemDate
						});

			const tableRow = await resolveKey(interaction, 'commands/check-boost-perks:tableRow', {
				amount: row.redeemableAmount,
				boostingWeeks: row.boostingWeeks,
				lastRedeemed,
				mention: row.mention,
				username: row.username,
				redeemableWeeks: row.redeemableWeeks
			});

			responseContainer.addTextDisplayComponents((textDisplay) => textDisplay.setContent(tableRow));
		}

		const summary = await resolveKey(interaction, 'commands/check-boost-perks:summary', {
			count: boosterRows.length
		});
		responseContainer.addTextDisplayComponents((textDisplay) => textDisplay.setContent(summary));

		return interaction.editReply({
			components: [responseContainer],
			flags: MessageFlags.IsComponentsV2
		});
	}
}
