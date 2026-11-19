/** @vitest-environment jsdom */
import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  createAdminStartup,
  getStartup,
  getStartupMarkets,
  listAdminStartupMarkets,
  listStartups,
  setAdminMarketStartup,
} from './startupsApi';

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe('startupsApi', () => {
  it('reads the public startup list envelope', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({
      ok: true,
      result: {
        startups: [{ id: 1, name: 'Step 3 Verified', slug: 'step3-verify' }],
        total: 1,
      },
    }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);

    const result = await listStartups();

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(String(fetchMock.mock.calls[0][0])).toContain('/v0/startups');
    expect(result.startups[0].slug).toBe('step3-verify');
  });

  it('reads one startup by slug', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({
      ok: true,
      result: {
        id: 1,
        name: 'Step 3 Verified',
        slug: 'step3-verify',
        description: 'Updated verification record',
        websiteUrl: 'https://example.com/verified',
      },
    }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);

    const result = await getStartup('step3-verify');

    expect(String(fetchMock.mock.calls[0][0])).toContain('/v0/startups/step3-verify');
    expect(result.name).toBe('Step 3 Verified');
    expect(result.websiteUrl).toBe('https://example.com/verified');
  });

  it('reads the public markets linked to a startup', async () => {
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({
      ok: true,
      result: {
        markets: [{ id: 11, questionTitle: 'Published question' }],
        total: 1,
      },
    }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);

    const result = await getStartupMarkets('step3-verify');

    expect(String(fetchMock.mock.calls[0][0])).toContain('/v0/startups/step3-verify/markets');
    expect(result.markets[0].id).toBe(11);
  });

  it('sends admin startup writes and market links with the signed-in token', async () => {
    window.localStorage.setItem('token', 'admin-token');
    const fetchMock = vi.fn(async () => new Response(JSON.stringify({
      ok: true,
      result: { id: 2, name: 'Northwind', slug: 'northwind' },
    }), { status: 201, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);

    await createAdminStartup({ name: 'Northwind', slug: 'northwind', description: '', websiteUrl: '' });
    await listAdminStartupMarkets();
    await setAdminMarketStartup(6, 'northwind');
    await setAdminMarketStartup(6, '');

    const calls = fetchMock.mock.calls.map(([url, options]) => ({
      url: String(url),
      method: options?.method || 'GET',
      body: options?.body,
      authorization: options?.headers?.Authorization,
    }));
    expect(calls[0].url).toContain('/v0/admin/startups');
    expect(calls[0].method).toBe('POST');
    expect(calls[0].authorization).toBe('Bearer admin-token');
    expect(calls[1].url).toContain('/v0/admin/startup-markets');
    expect(calls[2].url).toContain('/v0/admin/markets/6/startup');
    expect(calls[2].body).toContain('"startupSlug":"northwind"');
    expect(calls[3].body).toContain('"startupSlug":""');
    window.localStorage.removeItem('token');
  });
});
