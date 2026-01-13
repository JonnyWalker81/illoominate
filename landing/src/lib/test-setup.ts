import { vi, beforeEach, afterEach } from 'vitest';
import '@testing-library/jest-dom/vitest';

// Mock crypto.randomUUID for tests
if (!globalThis.crypto) {
  Object.defineProperty(globalThis, 'crypto', {
    value: {
      randomUUID: () => 'test-uuid-1234-5678-9012-abcdefghijkl',
      getRandomValues: (arr: Uint8Array) => {
        for (let i = 0; i < arr.length; i++) {
          arr[i] = Math.floor(Math.random() * 256);
        }
        return arr;
      },
      subtle: {} as SubtleCrypto,
    },
  });
}

// Store original fetch
const originalFetch = globalThis.fetch;

// Reset mocks between tests
beforeEach(() => {
  vi.clearAllMocks();
  // Reset fetch to mock by default
  globalThis.fetch = vi.fn();
});

afterEach(() => {
  // Restore original fetch after tests
  globalThis.fetch = originalFetch;
});
