import { FindingwayCommand } from '#lib/extensions/FindingwayComand';
import { ApplyOptions, RegisterChatInputCommand } from '@sapphire/decorators';
import { type ChatInputCommand } from '@sapphire/framework';
import { applyLocalizedBuilder, createLocalizedChoice } from '@sapphire/plugin-i18next';
import { ApplicationIntegrationType, MessageFlags } from 'discord.js';

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
							createLocalizedChoice('commands/settings:channelTypes.merc', { value: 'merc' }),
							createLocalizedChoice('commands/settings:channelTypes.uwu', { value: 'uwu' }),
							createLocalizedChoice('commands/settings:channelTypes.ucob', { value: 'ucob' }),
							createLocalizedChoice('commands/settings:channelTypes.tea', { value: 'tea' })
						)
				)
				.addChannelOption((builder) => applyLocalizedBuilder(builder, 'commands/settings:setChannel').setRequired(true))
		)
		.addSubcommand((builder) => applyLocalizedBuilder(builder, 'commands/settings:show'))
)
export class SlashCommands extends FindingwayCommand {
	public override async chatInputRun(interaction: ChatInputCommand.Interaction<'cached'>) {
		await interaction.deferReply({ flags: MessageFlags.Ephemeral });

		const subcommand = interaction.options.getSubcommand(true) as 'set' | 'show';

		switch (subcommand) {
			case 'set':
				return this.setSetting(interaction);
			case 'show':
				return this.showSettings(interaction);
		}
	}

	private async setSetting(interaction: ChatInputCommand.Interaction<'cached'>) {
		const type = interaction.options.getString('type', true) as 'merc' | 'uwu' | 'ucob' | 'tea';
		const channel = interaction.options.getChannel('channel', true);
	}

	private async showSettings(interaction: ChatInputCommand.Interaction<'cached'>) {
	}
}
