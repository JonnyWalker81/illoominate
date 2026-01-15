import type { Env } from '../../types.d.ts';

export function createMockEnv(overrides?: Partial<Env>): Env {
  return {
    DB: createMockD1(),
    TURNSTILE_SECRET_KEY: '1x0000000000000000000000000000000AA',
    RESEND_API_KEY: 're_test_key_placeholder',
    RESEND_SEGMENT_ID: 'seg_test_placeholder',
    RESEND_WEBHOOK_SECRET: 'whsec_test_placeholder',
    VERIFICATION_BASE_URL: 'http://localhost:4321',
    ADMIN_EMAIL: 'admin@test.com',
    FROM_EMAIL: 'noreply@test.com',
    ...overrides,
  };
}

export function createMockRequest(
  url: string,
  options: RequestInit = {}
): Request {
  const defaultHeaders = new Headers({
    'Content-Type': 'application/json',
    'CF-Connecting-IP': '127.0.0.1',
    'User-Agent': 'Test Agent',
  });

  if (options.headers) {
    const customHeaders = new Headers(options.headers);
    customHeaders.forEach((value, key) => {
      defaultHeaders.set(key, value);
    });
  }

  return new Request(url, {
    ...options,
    headers: defaultHeaders,
  });
}

interface MockD1Result {
  results: unknown[];
  success: boolean;
  meta: {
    duration: number;
    changes: number;
    last_row_id: number;
  };
}

interface MockD1PreparedStatement {
  bind: (...values: unknown[]) => MockD1PreparedStatement;
  first: <T = unknown>(column?: string) => Promise<T | null>;
  run: () => Promise<MockD1Result>;
  all: <T = unknown>() => Promise<{ results: T[]; success: boolean }>;
}

interface MockD1Database {
  prepare: (query: string) => MockD1PreparedStatement;
  batch: <T>(statements: MockD1PreparedStatement[]) => Promise<T[]>;
  exec: (query: string) => Promise<MockD1Result>;
  _data: Map<string, unknown[]>;
  _setData: (table: string, data: unknown[]) => void;
  _getData: (table: string) => unknown[];
}

export function createMockD1(): MockD1Database & D1Database {
  const data: Map<string, unknown[]> = new Map();

  const mockPreparedStatement = (query: string): MockD1PreparedStatement => {
    let boundValues: unknown[] = [];

    const statement: MockD1PreparedStatement = {
      bind(...values: unknown[]) {
        boundValues = values;
        return statement;
      },
      async first<T = unknown>(_column?: string): Promise<T | null> {
        // Simple mock implementation - returns first matching item
        const table = extractTableName(query);
        const tableData = data.get(table) || [];

        if (query.toLowerCase().includes('where')) {
          // Find matching record (simplified)
          const whereMatch = query.match(/where\s+(\w+)\s*=\s*\?/i);
          if (whereMatch) {
            const field = whereMatch[1];
            const value = boundValues[0];
            const found = tableData.find(
              (row) => (row as Record<string, unknown>)[field] === value
            );
            return (found as T) || null;
          }
        }

        return (tableData[0] as T) || null;
      },
      async run(): Promise<MockD1Result> {
        // Handle INSERT/UPDATE/DELETE
        const table = extractTableName(query);

        if (query.toLowerCase().startsWith('insert')) {
          const tableData = data.get(table) || [];
          const newId = tableData.length + 1;
          // Create a simple record with the bound values
          const record = { id: newId };
          tableData.push(record);
          data.set(table, tableData);
          return {
            results: [],
            success: true,
            meta: { duration: 0, changes: 1, last_row_id: newId },
          };
        }

        return {
          results: [],
          success: true,
          meta: { duration: 0, changes: 1, last_row_id: 0 },
        };
      },
      async all<T = unknown>(): Promise<{ results: T[]; success: boolean }> {
        const table = extractTableName(query);
        return {
          results: (data.get(table) || []) as T[],
          success: true,
        };
      },
    };

    return statement;
  };

  const mockDb: MockD1Database = {
    prepare: mockPreparedStatement,
    async batch<T>(statements: MockD1PreparedStatement[]): Promise<T[]> {
      const results: T[] = [];
      for (const stmt of statements) {
        const result = await stmt.run();
        results.push(result as unknown as T);
      }
      return results;
    },
    async exec(_query: string): Promise<MockD1Result> {
      return {
        results: [],
        success: true,
        meta: { duration: 0, changes: 0, last_row_id: 0 },
      };
    },
    _data: data,
    _setData(table: string, tableData: unknown[]) {
      data.set(table, tableData);
    },
    _getData(table: string) {
      return data.get(table) || [];
    },
  };

  return mockDb as MockD1Database & D1Database;
}

function extractTableName(query: string): string {
  // Extract table name from SQL query
  const fromMatch = query.match(/from\s+(\w+)/i);
  const intoMatch = query.match(/into\s+(\w+)/i);
  const updateMatch = query.match(/update\s+(\w+)/i);

  return fromMatch?.[1] || intoMatch?.[1] || updateMatch?.[1] || 'unknown';
}

export function mockFetch() {
  const originalFetch = globalThis.fetch;
  const calls: Array<{ url: string; options?: RequestInit }> = [];

  globalThis.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input.toString();
    calls.push({ url, options: init });

    // Mock Resend API responses
    if (url.includes('api.resend.com/emails')) {
      return new Response(JSON.stringify({ id: 'mock_email_id' }), {
        status: 200,
      });
    }

    if (url.includes('api.resend.com/contacts')) {
      return new Response(JSON.stringify({ id: 'mock_contact_id' }), {
        status: 200,
      });
    }

    return new Response('Not Found', { status: 404 });
  };

  return {
    restore() {
      globalThis.fetch = originalFetch;
    },
    getCalls() {
      return calls;
    },
  };
}
