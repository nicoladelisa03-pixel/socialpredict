/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import StartupProfile from './StartupProfile';
import { getStartup, getStartupMarkets } from '../../api/startupsApi';

vi.mock('../../api/startupsApi', () => ({
  getStartup: vi.fn(),
  getStartupMarkets: vi.fn(),
}));

const renderProfile = (path = '/startups/step3-verify') => render(
  <MemoryRouter initialEntries={[path]}>
    <Route path='/startups/:slug' component={StartupProfile} />
  </MemoryRouter>,
);

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

const emptyMarkets = () => {
  getStartupMarkets.mockResolvedValue({ markets: [], total: 0 });
};

describe('Startup profile', () => {
  it('shows the requested startup', async () => {
    emptyMarkets();
    let resolveStartup;
    getStartup.mockImplementation(() => new Promise((resolve) => {
      resolveStartup = resolve;
    }));

    renderProfile();

    expect(screen.getByRole('status').textContent).toContain('Loading startup...');
    resolveStartup({
      id: 1,
      name: 'Step 3 Verified',
      slug: 'step3-verify',
      description: 'Updated verification record',
      websiteUrl: 'https://example.com/verified',
    });

    expect(await screen.findByRole('heading', { name: 'Step 3 Verified' })).toBeTruthy();
    expect(screen.getByText('Updated verification record')).toBeTruthy();
    const website = screen.getByRole('link', { name: 'example.com' });
    expect(website.getAttribute('href')).toBe('https://example.com/verified');
    expect(website.getAttribute('target')).toBe('_blank');
    expect(website.getAttribute('rel')).toBe('noopener noreferrer');
    expect(screen.getByRole('link', { name: 'All startups' }).getAttribute('href')).toBe('/startups');
    expect(getStartup).toHaveBeenCalledWith('step3-verify');
  });

  it('shows linked markets and links them to the existing market pages', async () => {
    getStartup.mockResolvedValue({
      id: 1,
      name: 'Step 3 Verified',
      slug: 'step3-verify',
      description: 'Updated verification record',
    });
    let resolveMarkets;
    getStartupMarkets.mockImplementation(() => new Promise((resolve) => {
      resolveMarkets = resolve;
    }));

    renderProfile();

    expect(await screen.findByRole('heading', { name: 'Step 3 Verified' })).toBeTruthy();
    expect(screen.getByText('Loading markets...')).toBeTruthy();
    resolveMarkets({
      markets: [
        { id: 11, questionTitle: 'Published question', isResolved: false },
        { id: 12, questionTitle: 'Resolved question', isResolved: true, resolutionResult: 'YES' },
      ],
      total: 2,
    });

    const first = await screen.findByRole('link', { name: 'Published question' });
    const second = screen.getByRole('link', { name: 'Resolved question' });
    expect(first.getAttribute('href')).toBe('/markets/11');
    expect(second.getAttribute('href')).toBe('/markets/12');
    expect(screen.getByText('Resolved YES')).toBeTruthy();
  });

  it('shows an empty market list', async () => {
    emptyMarkets();
    getStartup.mockResolvedValue({ id: 1, name: 'Step 3 Verified', slug: 'step3-verify' });

    renderProfile();

    expect(await screen.findByText('No prediction markets are linked yet.')).toBeTruthy();
  });

  it('shows a market error and retries', async () => {
    getStartup.mockResolvedValue({ id: 1, name: 'Step 3 Verified', slug: 'step3-verify' });
    getStartupMarkets
      .mockRejectedValueOnce(new Error('Unable to load this startup\'s markets.'))
      .mockResolvedValueOnce({
        markets: [{ id: 11, questionTitle: 'Published question', isResolved: false }],
        total: 1,
      });

    renderProfile();

    expect((await screen.findByRole('alert')).textContent).toContain('Unable to load this startup\'s markets.');
    await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByRole('link', { name: 'Published question' })).toBeTruthy();
    expect(getStartupMarkets).toHaveBeenCalledTimes(2);
  });

  it('shows a fallback when the description is missing and hides an unsafe website', async () => {
    emptyMarkets();
    getStartup.mockResolvedValue({
      id: 2,
      name: 'No Site',
      slug: 'no-site',
      websiteUrl: 'javascript:alert(1)',
    });

    renderProfile('/startups/no-site');

    expect(await screen.findByText('No description has been added yet.')).toBeTruthy();
    expect(screen.getByText('No website has been added yet.')).toBeTruthy();
    expect(screen.queryByRole('link', { name: /alert/i })).toBeNull();
  });

  it('shows a not-found state for an unknown slug', async () => {
    emptyMarkets();
    const error = new Error('Not found');
    error.status = 404;
    getStartup.mockRejectedValue(error);

    renderProfile('/startups/missing-startup');

    expect(await screen.findByRole('heading', { name: 'Startup not found' })).toBeTruthy();
    expect(screen.getByText('No startup matches this address.')).toBeTruthy();
    expect(screen.getByRole('link', { name: 'All startups' }).getAttribute('href')).toBe('/startups');
  });

  it('shows an error and retries the request', async () => {
    emptyMarkets();
    getStartup
      .mockRejectedValueOnce(new Error('Unable to load this startup.'))
      .mockResolvedValueOnce({
        id: 1,
        name: 'Step 3 Verified',
        slug: 'step3-verify',
        description: 'Updated verification record',
      });

    renderProfile();

    expect((await screen.findByRole('alert')).textContent).toContain('Unable to load this startup.');
    await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByRole('heading', { name: 'Step 3 Verified' })).toBeTruthy();
    expect(getStartup).toHaveBeenCalledTimes(2);
  });
});