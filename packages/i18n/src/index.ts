/**
 * @solobueno/i18n - Internationalization Package
 *
 * Provides translation utilities and locale management.
 * Supports Spanish (es-419) and English (en) out of the box.
 *
 * @packageDocumentation
 */

import es419 from './locales/es-419.json';
import en from './locales/en.json';

export type Locale = 'es-419' | 'en';

const locales: Record<Locale, Record<string, string>> = {
  'es-419': es419,
  en: en,
};

let currentLocale: Locale = 'es-419';

/**
 * Set the current locale
 */
export function setLocale(locale: Locale): void {
  currentLocale = locale;
}

/**
 * Get the current locale
 */
export function getLocale(): Locale {
  return currentLocale;
}

/**
 * Translate a key to the current locale
 */
export function t(key: string, params?: Record<string, string>): string {
  const value = locales[currentLocale][key] || locales['en'][key] || key;

  if (!params) return value;

  return Object.entries(params).reduce(
    (acc, [k, v]) => acc.replace(`{{${k}}}`, v),
    value
  );
}

export { es419, en };
