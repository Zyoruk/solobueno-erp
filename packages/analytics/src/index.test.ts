import { describe, it, expect, beforeEach, vi } from 'vitest';
import { initAnalytics, trackEvent, trackScreen, trackAction, getQueuedEvents } from './index';

describe('@solobueno/analytics', () => {
  beforeEach(() => {
    // Clear console mock between tests
    vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  describe('initAnalytics', () => {
    it('should initialize without error', () => {
      expect(() => initAnalytics()).not.toThrow();
    });
  });

  describe('trackEvent', () => {
    it('should queue an event', () => {
      const initialCount = getQueuedEvents().length;
      trackEvent({ name: 'test_event' });
      expect(getQueuedEvents().length).toBeGreaterThan(initialCount);
    });

    it('should include timestamp in queued event', () => {
      trackEvent({ name: 'timestamp_test' });
      const events = getQueuedEvents();
      const lastEvent = events[events.length - 1];
      expect(lastEvent.timestamp).toBeDefined();
      expect(typeof lastEvent.timestamp).toBe('number');
    });

    it('should include properties in event', () => {
      trackEvent({
        name: 'props_test',
        properties: { foo: 'bar' },
      });
      const events = getQueuedEvents();
      const lastEvent = events[events.length - 1];
      expect(lastEvent.properties).toEqual({ foo: 'bar' });
    });
  });

  describe('trackScreen', () => {
    it('should track screen view event', () => {
      const initialCount = getQueuedEvents().length;
      trackScreen('HomeScreen');
      expect(getQueuedEvents().length).toBeGreaterThan(initialCount);
    });

    it('should include screen_name in properties', () => {
      trackScreen('TestScreen');
      const events = getQueuedEvents();
      const lastEvent = events[events.length - 1];
      expect(lastEvent.name).toBe('screen_view');
      expect(lastEvent.properties?.screen_name).toBe('TestScreen');
    });
  });

  describe('trackAction', () => {
    it('should track user action event', () => {
      trackAction('button_click');
      const events = getQueuedEvents();
      const lastEvent = events[events.length - 1];
      expect(lastEvent.name).toBe('user_action');
      expect(lastEvent.properties?.action).toBe('button_click');
    });
  });

  describe('getQueuedEvents', () => {
    it('should return array of events', () => {
      const events = getQueuedEvents();
      expect(Array.isArray(events)).toBe(true);
    });

    it('should return copy of queue (immutable)', () => {
      const events1 = getQueuedEvents();
      const events2 = getQueuedEvents();
      expect(events1).not.toBe(events2);
    });
  });
});
