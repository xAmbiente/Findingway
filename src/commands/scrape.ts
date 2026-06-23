import { FindingwayCommand } from '#lib/extensions/FindingwayComand';
import { ApplyOptions, RegisterChatInputCommand } from '@sapphire/decorators';
import { type ChatInputCommand } from '@sapphire/framework';
import { applyLocalizedBuilder } from '@sapphire/plugin-i18next';
import { ApplicationIntegrationType, MessageFlags } from 'discord.js';

@ApplyOptions<ChatInputCommand.Options>({
	description: 'Triggers an immediate scrape'
})
@RegisterChatInputCommand((builder) =>
	applyLocalizedBuilder(builder, 'commands/scrape:root').setIntegrationTypes(ApplicationIntegrationType.GuildInstall)
)
export class SlashCommands extends FindingwayCommand {
	public override async chatInputRun(interaction: ChatInputCommand.Interaction<'cached'>) {
		await interaction.deferReply({ flags: MessageFlags.Ephemeral });

		return interaction.editReply('Scraping...');
	}
}
