import { envParseString } from '@skyra/env-utilities';

export const BOOST_PERK_REDEMPTION_CHANNEL_ID = envParseString('BOOST_PERK_REDEMPTION_CHANNEL_ID');
export const BOOST_PERK_WEEK_MILLISECONDS = 7 * 24 * 60 * 60 * 1_000;
export const BOOST_PERK_WEEKLY_GIL = 200_000;
export const BOOST_PERK_REDEEM_BUTTON_CUSTOM_ID = 'boost-perk:redeem-now';
export const BOOST_PERK_PAYOUT_COORDINATOR_ID = '1258172662254141603'; // Ambi
