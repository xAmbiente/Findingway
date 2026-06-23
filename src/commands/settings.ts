import { FindingwayCommand } from '#lib/extensions/FindingwayComand';
import { ChannelType } from '#lib/generated/prisma-client/enums';
import { BrandingColors, ErrorIdentifiers } from '#utils/constants';
import { FindingwayEmojis } from '#utils/emojis';
import { ApplyOptions, RegisterChatInputCommand } from '@sapphire/decorators';
import { UserError, type ChatInputCommand } from '@sapphire/framework';
import { applyLocalizedBuilder, createLocalizedChoice, resolveKey } from '@sapphire/plugin-i18next';
import { ApplicationIntegrationType, bold, channelMention, ContainerBuilder, MessageFlags } from 'discord.js';

@ApplyOptions<ChatInputCommand.Options>({
	description: 'Change the settings of the bot'
})
@RegisterChatInputCommand((builder) =>
	applyLocalizedBuilder(builder, 'commands/settings:root')
		.setIntegrationTypes(ApplicationIntegrationType.GuildInstall)
		.addSubcommand((builder) =>
			applyLocalizedBuilder(builder, 'commands/settings:set') //
				.addStringOption((builder) =>
					applyLocalizedBuilder(builder, 'commands/settings:setType')
						.setRequired(true)
						.setChoices(
							createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.Mercantile}`, { value: ChannelType.Mercantile }),
							createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.TheWeaponsRefrain}`, {
								value: ChannelType.TheWeaponsRefrain
							}),
							createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.TheUnendingCoilOfBahamut}`, {
								value: ChannelType.TheUnendingCoilOfBahamut
							}),
							createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.TheEpicOfAlexander}`, {
								value: ChannelType.TheEpicOfAlexander
							})
						)
				)
				.addChannelOption((builder) => applyLocalizedBuilder(builder, 'commands/settings:setChannel').setRequired(true))
		)
		.addSubcommand((builder) =>
			applyLocalizedBuilder(builder, 'commands/settings:toggle').addStringOption((builder) =>
				applyLocalizedBuilder(builder, 'commands/settings:setType')
					.setRequired(true)
					.setChoices(
						createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.Mercantile}`, { value: ChannelType.Mercantile }),
						createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.TheWeaponsRefrain}`, {
							value: ChannelType.TheWeaponsRefrain
						}),
						createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.TheUnendingCoilOfBahamut}`, {
							value: ChannelType.TheUnendingCoilOfBahamut
						}),
						createLocalizedChoice(`commands/settings:channelTypes.${ChannelType.TheEpicOfAlexander}`, {
							value: ChannelType.TheEpicOfAlexander
						})
					)
			)
		)
		.addSubcommand((builder) => applyLocalizedBuilder(builder, 'commands/settings:show'))
)
export class SlashCommands extends FindingwayCommand {
	public override async chatInputRun(interaction: ChatInputCommand.Interaction<'cached'>) {
		await interaction.deferReply({ flags: MessageFlags.Ephemeral });

		const subcommand = interaction.options.getSubcommand(true) as 'set' | 'show' | 'toggle';

		switch (subcommand) {
			case 'set':
				return this.setSetting(interaction);
			case 'toggle':
				return this.toggleSetting(interaction);
			case 'show':
				return this.showSettings(interaction);
		}
	}

	private async setSetting(interaction: ChatInputCommand.Interaction<'cached'>) {
		const type = interaction.options.getString('type', true) as ChannelType;
		const channel = interaction.options.getChannel('channel', true);

		await this.container.prisma.channel.upsert({
			create: {
				channelId: channel.id,
				guildId: interaction.guildId,
				type
			},
			update: {
				channelId: channel.id,
				guildId: interaction.guildId,
				type
			},
			where: {
				type_guildId: {
					type,
					guildId: interaction.guildId
				}
			}
		});

		return interaction.editReply({
			content: await resolveKey(interaction, 'commands/settings:setSuccessful', {
				channel: channelMention(channel.id),
				type: await resolveKey(interaction, `commands/settings:channelTypes.${type}`)
			})
		});
	}

	private async toggleSetting(interaction: ChatInputCommand.Interaction<'cached'>) {
		const type = interaction.options.getString('type', true) as ChannelType;

		const channelToModify = await this.container.prisma.channel.findFirst({
			where: {
				type,
				guildId: interaction.guildId
			}
		});

		if (!channelToModify) {
			throw new UserError({
				message: await resolveKey(interaction, 'commands/settings:toggleNoChannelConfigured', {
					type: await resolveKey(interaction, `commands/settings:channelTypes.${type}`)
				}),
				identifier: ErrorIdentifiers.SetNoChannelConfigured
			});
		}

		await this.container.prisma.channel.update({
			data: {
				enabled: !channelToModify.enabled
			},
			where: {
				type_guildId: {
					type,
					guildId: interaction.guildId
				}
			}
		});

		return interaction.editReply({
			content: await resolveKey(interaction, 'commands/settings:toggleSuccessful', {
				state: channelToModify.enabled
					? await resolveKey(interaction, 'commands/settings:disabled')
					: await resolveKey(interaction, 'commands/settings:enabled'),
				type: await resolveKey(interaction, `commands/settings:channelTypes.${type}`)
			})
		});
	}

	private async showSettings(interaction: ChatInputCommand.Interaction<'cached'>) {
		const channels = await this.container.prisma.channel.findMany({
			where: {
				guildId: interaction.guildId
			}
		});

		if (channels.length === 0) {
			throw new UserError({
				message: await resolveKey(interaction, 'commands/settings:noSettings'),
				identifier: ErrorIdentifiers.SetNoChannelConfigured
			});
		}

		const settings = await Promise.all(
			channels.map(async (channel) => {
				const channelTypeName = await resolveKey(interaction, `commands/settings:channelTypes.${channel.type}`);
				const statusEmoji = channel.enabled ? FindingwayEmojis.GreenTick : FindingwayEmojis.RedCross;
				return {
					name: channelTypeName,
					value: `${statusEmoji} ${bold(channelTypeName)} - ${channelMention(channel.channelId)} [DC: Light]`
				};
			})
		);

		const raidContent = settings.map((setting) => setting.value);

		const container = new ContainerBuilder() //
			.setAccentColor(BrandingColors.Primary)
			.addTextDisplayComponents((textDisplay) => textDisplay.setContent(raidContent.join('\n')));

		return interaction.editReply({
			components: [container],
			flags: MessageFlags.IsComponentsV2
		});
	}
}
