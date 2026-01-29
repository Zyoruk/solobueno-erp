/**
 * @solobueno/analytics - Analytics Tracking
 *
 * Utilities for tracking user behavior and business events.
 * Supports offline queuing and multiple analytics providers.
 *
 * @packageDocumentation
 */

export interface AnalyticsEvent {
  name: string;
  properties?: Record<string, unknown>;
  timestamp?: number;
}

interface QueuedEvent extends AnalyticsEvent {
  timestamp: number;
  id: string;
}

const eventQueue: QueuedEvent[] = [];
let isInitialized = false;

/**
 * Initialize the analytics system
 */
export function initAnalytics(): void {
  isInitialized = true;
  // TODO: Process queued events
}

/**
 * Track an analytics event
 */
export function trackEvent(event: AnalyticsEvent): void {
  const queuedEvent: QueuedEvent = {
    ...event,
    timestamp: event.timestamp || Date.now(),
    id: generateId(),
  };

  eventQueue.push(queuedEvent);

  if (isInitialized) {
    // TODO: Send to analytics service
    console.log('[Analytics]', queuedEvent);
  }
}

/**
 * Track a page/screen view
 */
export function trackScreen(screenName: string, properties?: Record<string, unknown>): void {
  trackEvent({
    name: 'screen_view',
    properties: {
      screen_name: screenName,
      ...properties,
    },
  });
}

/**
 * Track a user action
 */
export function trackAction(action: string, properties?: Record<string, unknown>): void {
  trackEvent({
    name: 'user_action',
    properties: {
      action,
      ...properties,
    },
  });
}

/**
 * Get queued events (for debugging or manual sync)
 */
export function getQueuedEvents(): QueuedEvent[] {
  return [...eventQueue];
}

function generateId(): string {
  return Math.random().toString(36).substring(2, 15);
}
