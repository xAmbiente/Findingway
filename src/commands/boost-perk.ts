import { BrandingColors } from '#utils/constants';
import { BOOST_PERK_PAYOUT_COORDINATOR_ID, BOOST_PERK_REDEEM_BUTTON_CUSTOM_ID } from '#utils/constants/boost-perk';
import { getBoostPerkRedeemStatus, runBoostPerkRedeemFlow } from '#utils/functions/boostPerk';
import { ApplyOptions, RegisterChatInputCommand } from '@sapphire/decorators';
import { isGuildMember } from '@sapphire/discord.js-utilities';
import { Command, type ChatInputCommand } from '@sapphire/framework';
import { applyLocalizedBuilder, resolveKey } from '@sapphire/plugin-i18next';
import { ApplicationIntegrationType, ButtonBuilder, ButtonStyle, ContainerBuilder, MessageFlags, PermissionFlagsBits } from 'discord.js';

@ApplyOptions<ChatInputCommand.Options>({
	requiredClientPermissions: [PermissionFlagsBits.CreatePrivateThreads]
})
@RegisterChatInputCommand((builder) =>
	applyLocalizedBuilder(builder, 'commands/boost-perk:root')
		.setIntegrationTypes(ApplicationIntegrationType.GuildInstall)
		.addSubcommand((builder) => applyLocalizedBuilder(builder, 'commands/boost-perk:check'))
		.addSubcommand((builder) => applyLocalizedBuilder(builder, 'commands/boost-perk:redeem'))
)
export class SlashCommand extends Command {
	public override async chatInputRun(interaction: ChatInputCommand.Interaction) {
		await interaction.deferReply({ flags: MessageFlags.Ephemeral });

		if (!isGuildMember(interaction.member)) {
			return interaction.editReply({
				content: await resolveKey(interaction, 'commands/boost-perk:genericMemberError')
			});
		}

		const subcommand = interaction.options.getSubcommand(true) as 'check' | 'redeem';

		if (interaction.user.id === BOOST_PERK_PAYOUT_COORDINATOR_ID) {
			const messageKey = subcommand === 'check' ? 'commands/boost-perk:checkNotEligible' : 'commands/boost-perk:redeemNotEligible';

			return interaction.editReply({
				content: await resolveKey(interaction, messageKey)
			});
		}

		switch (subcommand) {
			case 'check':
				return this.check(interaction as ChatInputCommand.Interaction<'cached'>);
			case 'redeem':
				return this.redeem(interaction as ChatInputCommand.Interaction<'cached'>);
		}
	}

	private async check(interaction: ChatInputCommand.Interaction<'cached'>) {
		const status = await getBoostPerkRedeemStatus(interaction);
		const contentRows: string[] = [await resolveKey(interaction, 'commands/boost-perk:checkHeader')];

		if (status.isPremiumMember) {
			contentRows.push(
				await resolveKey(interaction, 'commands/boost-perk:checkPremiumSince', {
					premiumSince: status.premiumSince
				}),
				await resolveKey(interaction, 'commands/boost-perk:checkWeeksPremium', {
					weeksPremium: status.weeksPremium
				}),
				status.lastRedeemDate
					? await resolveKey(interaction, 'commands/boost-perk:checkLastRedeemed', {
							lastRedeemDate: status.lastRedeemDate
						})
					: await resolveKey(interaction, 'commands/boost-perk:checkLastRedeemedNever'),
				await resolveKey(interaction, 'commands/boost-perk:checkRedeemableWeeks', {
					redeemableWeeks: status.redeemableWeeks
				}),
				await resolveKey(interaction, 'commands/boost-perk:checkRedeemableAmount', {
					amount: status.redeemableAmount
				}),
				status.eligible
					? await resolveKey(interaction, 'commands/boost-perk:checkEligible')
					: await resolveKey(interaction, 'commands/boost-perk:checkNotEligible')
			);
		} else {
			contentRows.push(await resolveKey(interaction, 'commands/boost-perk:checkNotPremium'));
		}

		const container = new ContainerBuilder() //
			.setAccentColor(BrandingColors.Primary)
			.addTextDisplayComponents((textDisplay) => textDisplay.setContent(contentRows.join('\n')));

		if (status.eligible) {
			const redeemNowButtonLabel = await resolveKey(interaction, 'commands/boost-perk:redeemNowButton');

			container.addActionRowComponents((row) =>
				row.addComponents(
					new ButtonBuilder().setCustomId(BOOST_PERK_REDEEM_BUTTON_CUSTOM_ID).setStyle(ButtonStyle.Success).setLabel(redeemNowButtonLabel)
				)
			);
		}

		await interaction.editReply({
			components: [container],
			flags: MessageFlags.IsComponentsV2
		});
	}

	private async redeem(interaction: ChatInputCommand.Interaction<'cached'>) {
		return runBoostPerkRedeemFlow(interaction);
	}
}
