import { Command } from '@sapphire/framework';
import { PermissionFlagsBits } from 'discord.js';

export abstract class FindingwayCommand extends Command {
	public constructor(context: Command.LoaderContext, options: Command.Options) {
		super(context, {
			requiredClientPermissions: [PermissionFlagsBits.EmbedLinks],
			requiredUserPermissions: [PermissionFlagsBits.ManageMessages],
			...options
		});
	}
}
