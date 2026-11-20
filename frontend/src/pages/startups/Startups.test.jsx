/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Startups from './Startups';
import { listStartups } from '../../api/startupsApi';

const renderDirectory = () => render(
  <MemoryRouter>
    <Startups />
  </MemoryRouter>,
);

vi.mock('../../api/startupsApi', () => ({
  listStartups: vi.fn(),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('Startups directory', () => {
  it('shows startups returned by the API', async () => {
    let resolveList;
    listStartups.mockImplementation(() => new Promise((resolve) => {
      resolveList = resolve;
    }));

    renderDirectory();

    expect(screen.getByRole('status').textContent).toContain('Loading startups...');
    resolveList({
      startups: [
        {
          id: 1,
          name: 'Step 3 Verified',
          slug: 'step3-verify',
          description: 'Updated verification record',
          websiteUrl: 'https://example.com/verified',
        },
        {
          id: 2,
          name: 'No Site',
          slug: 'no-site',
          websiteUrl: 'javascript:alert(1)',
        },
      ],
      total: 2,
    });
    expect(await screen.findByRole('heading', { name: 'Step 3 Verified' })).toBeTruthy();
    expect(screen.getByText('Updated verification record')).toBeTruthy();
    const website = screen.getByRole('link', { name: 'example.com' });
    expect(website.getAttribute('href')).toBe('https://example.com/verified');
    expect(website.getAttribute('target')).toBe('_blank');
    expect(website.getAttribute('rel')).toBe('noopener noreferrer');
    expect(screen.getByRole('heading', { name: 'No Site' })).toBeTruthy();
    expect(screen.queryByRole('link', { name: /alert/i })).toBeNull();
    expect(screen.getByRole('link', { name: 'Step 3 Verified' }).getAttribute('href')).toBe('/startups/step3-verify');
    expect(screen.getByRole('link', { name: 'No Site' }).getAttribute('href')).toBe('/startups/no-site');
    expect(screen.getByText('Test record')).toBeTruthy();
  });

  it('shows an empty state when the API returns no startups', async () => {
    listStartups.mockResolvedValue({ startups: [], total: 0 });

    renderDirectory();

    expect(await screen.findByText('No startups have been added yet.')).toBeTruthy();
  });

  it('shows an error and retries the request', async () => {
    listStartups
      .mockRejectedValueOnce(new Error('Unable to load startups.'))
      .mockResolvedValueOnce({
        startups: [{ id: 1, name: 'Step 3 Verified', slug: 'step3-verify' }],
        total: 1,
      });

    renderDirectory();

    expect((await screen.findByRole('alert')).textContent).toContain('Unable to load startups.');
    await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(await screen.findByRole('heading', { name: 'Step 3 Verified' })).toBeTruthy();
    expect(listStartups).toHaveBeenCalledTimes(2);
  });
});
