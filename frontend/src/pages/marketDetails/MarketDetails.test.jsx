/** @vitest-environment jsdom */
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import MarketDetails from './MarketDetails';

const renderMarket = () => render(
  <MemoryRouter initialEntries={['/markets/41']}>
    <Route path='/markets/:marketId' component={MarketDetails} />
  </MemoryRouter>,
);

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('Market details availability', () => {
  it('shows not available when the detail API refuses the market', async () => {
    vi.stubGlobal('fetch', vi.fn(async (url) => ({
      ok: false,
      status: String(url).includes('/summary') ? 404 : 404,
      json: async () => ({ ok: false, reason: 'MARKET_NOT_FOUND' }),
    })));

    renderMarket();

    await waitFor(() => {
      expect(screen.getByRole('status').textContent).toContain('Market not found');
    });
    expect(screen.getByRole('status').textContent).toContain('This market is not available.');
    expect(document.body.textContent).not.toContain('Secret proposed title');
  });
});
