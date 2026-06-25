import { FindingwayCommand } from '#lib/extensions/FindingwayComand';
import { scrape } from '#lib/scraper/xivpfScraper';
import { FindingwayEvents } from '#utils/constants';
import { ApplyOptions, RegisterChatInputCommand } from '@sapphire/decorators';
import { type ChatInputCommand } from '@sapphire/framework';
import { applyLocalizedBuilder, resolveKey } from '@sapphire/plugin-i18next';
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

		const timeOfScrape = Date.now();
		const scrapeJson = await scrape();

		this.container.client.emit(FindingwayEvents.PostTea, { entries: scrapeJson.TheEpicOfAlexander, timeOfScrape });
		this.container.client.emit(FindingwayEvents.PostUcob, { entries: scrapeJson.TheUnendingCoilOfBahamut, timeOfScrape });
		this.container.client.emit(FindingwayEvents.PostUwu, { entries: scrapeJson.TheWeaponsRefrain, timeOfScrape });

		const response = await resolveKey(interaction, 'commands/scrape:success');

		return interaction.editReply(response);
	}
}
