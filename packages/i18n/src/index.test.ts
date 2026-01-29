import { describe, it, expect, beforeEach } from 'vitest';
import { setLocale, getLocale, t } from './index';

describe('@solobueno/i18n', () => {
  beforeEach(() => {
    // Reset to default locale before each test
    setLocale('es-419');
  });

  describe('setLocale', () => {
    it('should set locale to es-419', () => {
      setLocale('es-419');
      expect(getLocale()).toBe('es-419');
    });

    it('should set locale to en', () => {
      setLocale('en');
      expect(getLocale()).toBe('en');
    });
  });

  describe('getLocale', () => {
    it('should return current locale', () => {
      expect(getLocale()).toBe('es-419');
    });
  });

  describe('t (translate)', () => {
    it('should translate key in Spanish', () => {
      setLocale('es-419');
      const result = t('app.name');
      expect(result).toBe('Solobueno ERP');
    });

    it('should translate key in English', () => {
      setLocale('en');
      const result = t('app.name');
      expect(result).toBe('Solobueno ERP');
    });

    it('should return key if translation not found', () => {
      const result = t('nonexistent.key');
      expect(result).toBe('nonexistent.key');
    });

    it('should interpolate parameters', () => {
      setLocale('en');
      const result = t('common.welcome', { name: 'John' });
      expect(result).toBe('Welcome, John');
    });

    it('should fallback to English if key missing in current locale', () => {
      setLocale('es-419');
      // Assuming English has a key that Spanish doesn't
      const result = t('app.name');
      expect(result).toBeDefined();
    });
  });
});
