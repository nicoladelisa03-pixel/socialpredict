/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import StartupProfile from './StartupProfile';
import { getStartup, getStartupFunding, getStartupMarkets } from '../../api/startupsApi';

vi.mock('../../api/startupsApi', () => ({
  getStartup: vi.fn(),
  getStartupMarkets: vi.fn(),
  getStartupFunding: vi.fn(async () => ({ rounds: [], latestReportedValuation: null })),
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

  it('shows sourced funding in announcement order and the latest reported valuation', async () => {
    emptyMarkets();
    getStartup.mockResolvedValue({
      id: 8,
      name: 'Mercor',
      slug: 'mercor',
      description: 'A private company',
    });
    getStartupFunding.mockResolvedValueOnce({
      rounds: [
        {
          id: 1,
          roundLabel: 'Seed',
          announcementDate: '2024-01-02',
          amountRaisedMinor: 125050,
          valuationMinor: null,
          valuationBasis: 'unspecified',
          currency: 'USD',
          evidenceUrl: 'https://example.com/seed',
          sourceLabel: 'Seed announcement',
        },
        {
          id: 2,
          roundLabel: 'Series A',
          announcementDate: '2025-06-01',
          amountRaisedMinor: null,
          valuationMinor: 900000,
          valuationBasis: 'post-money',
          currency: 'USD',
          evidenceUrl: 'https://example.com/series-a',
          sourceLabel: 'Series A announcement',
        },
      ],
      latestReportedValuation: {
        id: 2,
        roundLabel: 'Series A',
        announcementDate: '2025-06-01',
        valuationMinor: 900000,
        valuationBasis: 'post-money',
        currency: 'USD',
        evidenceUrl: 'https://example.com/series-a',
        sourceLabel: 'Series A announcement',
      },
    });

    renderProfile('/startups/mercor');

    expect(await screen.findByRole('heading', { name: 'Latest reported valuation' })).toBeTruthy();
    expect(screen.getAllByText('USD 9,000.00').length).toBeGreaterThan(0);
    expect(screen.getByText(/Reported on 2025-06-01\. post-money\./)).toBeTruthy();
    expect(screen.getByText(/not a prediction-market price or a probability/)).toBeTruthy();
    const source = screen.getByRole('link', { name: 'Source: Series A announcement' });
    expect(source.getAttribute('href')).toBe('https://example.com/series-a');
    const headings = screen.getAllByRole('heading', { name: /Seed|Series A/ }).map((heading) => heading.textContent);
    expect(headings).toEqual(['Seed', 'Series A']);
    expect(screen.getByText('Amount raised USD 1,250.50')).toBeTruthy();
    expect(screen.getByText('Amount raised was not disclosed')).toBeTruthy();
    expect(screen.getByText('Reported valuation was not disclosed')).toBeTruthy();
    expect(screen.getByText('Reported valuation USD 9,000.00 (post-money)')).toBeTruthy();
  });

  it('shows Mercor Series D as a sourced valuation, separate from market prices', async () => {
    emptyMarkets();
    getStartup.mockResolvedValue({
      id: 12,
      name: 'Mercor',
      slug: 'mercor',
      description: 'Mercor organizes human expertise for AI work.',
    });
    getStartupFunding.mockResolvedValueOnce({
      rounds: [{
        id: 1,
        roundLabel: 'Series D',
        announcementDate: '2026-11-10',
        amountRaisedMinor: 75000000000,
        valuationMinor: 2000000000000,
        valuationBasis: 'unspecified',
        currency: 'USD',
        evidenceUrl: 'https://www.mercor.com/blog/series-d/',
        sourceLabel: 'Mercor',
      }],
      latestReportedValuation: {
        id: 1,
        roundLabel: 'Series D',
        announcementDate: '2026-11-10',
        valuationMinor: 2000000000000,
        valuationBasis: 'unspecified',
        currency: 'USD',
        evidenceUrl: 'https://www.mercor.com/blog/series-d/',
        sourceLabel: 'Mercor',
      },
    });

    renderProfile('/startups/mercor');

    expect(await screen.findByText('USD 20,000,000,000.00')).toBeTruthy();
    expect(screen.getByText('Amount raised USD 750,000,000.00')).toBeTruthy();
    expect(screen.getByText(/Reported on 2026-11-10\. basis not stated\./)).toBeTruthy();
    expect(screen.getByText(/not a prediction-market price or a probability/)).toBeTruthy();
    const source = screen.getByRole('link', { name: 'Source: Mercor' });
    expect(source.getAttribute('href')).toBe('https://www.mercor.com/blog/series-d/');
    expect(screen.queryByText(/probability/i).textContent).not.toMatch(/YES|NO/);
  });

  it('shows the empty funding state without inventing a figure', async () => {
    emptyMarkets();
    getStartup.mockResolvedValue({ id: 1, name: 'Step 3 Verified', slug: 'step3-verify' });

    renderProfile();

    expect(await screen.findByText('No verified funding data added yet')).toBeTruthy();
    expect(screen.queryByText(/USD /)).toBeNull();
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