/** @vitest-environment jsdom */
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import AdminStartups from './AdminStartups';
import {
  createAdminStartup,
  listAdminStartupMarkets,
  listStartups,
  setAdminMarketStartup,
  updateAdminStartup,
} from '../../../api/startupsApi';

vi.mock('../../../api/startupsApi', () => ({
  listStartups: vi.fn(),
  createAdminStartup: vi.fn(),
  updateAdminStartup: vi.fn(),
  listAdminStartupMarkets: vi.fn(),
  setAdminMarketStartup: vi.fn(),
}));

const publishedMarket = {
  marketId: 6,
  questionTitle: 'Visible approved market',
  lifecycleStatus: 'published',
  startupSlug: '',
};

const proposedMarket = {
  marketId: 7,
  questionTitle: 'Hidden proposed market',
  lifecycleStatus: 'proposed',
  startupSlug: '',
};

const renderAdmin = () => render(
  <MemoryRouter>
    <AdminStartups />
  </MemoryRouter>,
);

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('Admin startups', () => {
  it('creates, edits, links, and clears a market from the admin screen', async () => {
    let startups = [];
    let markets = [{ ...publishedMarket }, { ...proposedMarket }];
    listStartups.mockImplementation(async () => ({ startups: startups.map((startup) => ({ ...startup })), total: startups.length }));
    listAdminStartupMarkets.mockImplementation(async () => ({ markets: markets.map((market) => ({ ...market })), total: markets.length }));
    createAdminStartup.mockImplementation(async (body) => {
      const created = { id: 3, ...body };
      startups = [...startups, created];
      return created;
    });
    updateAdminStartup.mockImplementation(async (slug, body) => {
      startups = startups.map((startup) => (startup.slug === slug ? { ...startup, ...body, slug } : startup));
      return startups.find((startup) => startup.slug === slug);
    });
    setAdminMarketStartup.mockImplementation(async (marketId, startupSlug) => {
      markets = markets.map((market) => (
        market.marketId === marketId ? { ...market, startupSlug } : market
      ));
      return { marketId, startupSlug };
    });

    renderAdmin();
    expect(await screen.findByText('No startups yet. Create one to link markets to it.')).toBeTruthy();

    await userEvent.type(screen.getByLabelText('Name'), 'Northwind');
    await userEvent.type(screen.getByLabelText('Slug'), 'northwind');
    await userEvent.type(screen.getByLabelText('Description'), 'A test company');
    await userEvent.type(screen.getByLabelText('Website'), 'https://northwind.example');
    await userEvent.click(screen.getByRole('button', { name: 'Create startup' }));

    expect(await screen.findByRole('button', { name: /Northwind/ })).toBeTruthy();
    expect(createAdminStartup).toHaveBeenCalledWith({
      name: 'Northwind',
      slug: 'northwind',
      description: 'A test company',
      websiteUrl: 'https://northwind.example',
    });
    expect(document.querySelector('span.font-mono').textContent).toBe('northwind');

    const editName = document.getElementById('edit-startup-name');
    await userEvent.clear(editName);
    await userEvent.type(editName, 'Northwind Labs');
    await userEvent.click(screen.getByRole('button', { name: 'Save startup' }));
    await waitFor(() => {
      expect(updateAdminStartup).toHaveBeenCalledWith('northwind', {
        name: 'Northwind Labs',
        description: 'A test company',
        websiteUrl: 'https://northwind.example',
      });
    });
    expect(await screen.findByText('Startup saved.')).toBeTruthy();

    await userEvent.selectOptions(screen.getByLabelText('Market'), '6');
    expect(screen.getByText('Current link: not linked')).toBeTruthy();
    await userEvent.selectOptions(screen.getByLabelText('Startup'), 'northwind');
    await userEvent.click(screen.getByRole('button', { name: 'Save market link' }));
    await waitFor(() => {
      expect(screen.getByText('Current link: Northwind Labs (northwind)')).toBeTruthy();
    });
    expect(setAdminMarketStartup).toHaveBeenCalledWith(6, 'northwind');
    expect(screen.getByRole('button', { name: 'Change link' })).toBeTruthy();

    await userEvent.selectOptions(screen.getByLabelText('Startup'), '');
    await userEvent.click(screen.getByRole('button', { name: 'Save market link' }));
    await waitFor(() => {
      expect(screen.getByText('Current link: not linked')).toBeTruthy();
    });
    expect(setAdminMarketStartup).toHaveBeenLastCalledWith(6, '');
    expect(screen.getByRole('option', { name: '#7 Hidden proposed market' })).toBeTruthy();
  });

  it('shows validation, permission, conflict, and load failures', async () => {
    listStartups.mockResolvedValue({ startups: [], total: 0 });
    listAdminStartupMarkets.mockResolvedValue({ markets: [], total: 0 });
    renderAdmin();
    await screen.findByText('No startups yet. Create one to link markets to it.');

    await userEvent.type(screen.getByLabelText('Name'), 'Bad');
    await userEvent.type(screen.getByLabelText('Slug'), 'bad slug');
    await userEvent.click(screen.getByRole('button', { name: 'Create startup' }));
    expect(screen.getByRole('alert').textContent).toContain('lowercase letters');
    expect(createAdminStartup).not.toHaveBeenCalled();

    await userEvent.clear(screen.getByLabelText('Slug'));
    await userEvent.type(screen.getByLabelText('Slug'), 'bad-slug');
    createAdminStartup.mockRejectedValueOnce(Object.assign(new Error('A startup with this slug already exists.'), { status: 409 }));
    await userEvent.click(screen.getByRole('button', { name: 'Create startup' }));
    expect(await screen.findByRole('alert')).toHaveProperty('textContent', expect.stringContaining('already exists'));

    cleanup();
    listAdminStartupMarkets.mockRejectedValueOnce(Object.assign(new Error('denied'), { status: 403 }));
    renderAdmin();
    expect((await screen.findByRole('alert')).textContent).toContain('admin account');

    cleanup();
    listStartups.mockRejectedValueOnce(Object.assign(new Error('missing token'), { status: 401 }));
    listAdminStartupMarkets.mockResolvedValue({ markets: [], total: 0 });
    renderAdmin();
    expect((await screen.findByRole('alert')).textContent).toContain('Sign in again');

    cleanup();
    listStartups.mockRejectedValueOnce(new Error('network down'));
    listAdminStartupMarkets.mockResolvedValue({ markets: [], total: 0 });
    renderAdmin();
    expect((await screen.findByRole('alert')).textContent).toContain('network down');
    listStartups.mockResolvedValue({ startups: [], total: 0 });
    await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByText('No startups yet. Create one to link markets to it.')).toBeTruthy();
  });
});
