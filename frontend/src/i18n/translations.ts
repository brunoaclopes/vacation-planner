import { en, ptPT } from './locales';
import { Language, Translations } from './types';

export type { Language, Translations };

export const translations: Record<Language, Translations> = {
  en,
  'pt-PT': ptPT,
};
