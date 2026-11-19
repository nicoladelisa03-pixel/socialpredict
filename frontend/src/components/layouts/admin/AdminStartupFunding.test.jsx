/** @vitest-environment jsdom */
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import AdminStartupFunding from './AdminStartupFunding';
import {
  createAdminStartupFunding,
  getStartupFunding,
  updateAdminStartupFunding,
} from '../../../api/startupsApi';

vi.mock('../../../api/startupsApi', () => ({
  getStartupFunding: vi.fn(),
  createAdminStartupFunding: vi.fn(),
  updateAdminStartupFunding: vi.fn(),
}));

const renderFunding = () => render(<AdminStartupFunding slug='mercor' />);

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('Admin startup funding', () => {
  it('adds a funding round and keeps an unknown valuation null', async () => {
    let rounds = [];
    getStartupFunding.mockImplementation(async () => ({ rounds: rounds.map((round) => ({ ...round })), latestReportedValuation: null }));
    createAdminStartupFunding.mockImplementation(async (_slug, body) => {
      const created = { id: 4, ...body };
      rounds = [...rounds, created];
      return created;
    });

    renderFunding();
    expect(await screen.findByText('No verified funding data added yet')).toBeTruthy();

    await userEvent.type(screen.getByLabelText('Round label'), 'Seed');
    await userEvent.type(screen.getByLabelText('Announcement date'), '2024-01-02');
    await userEvent.type(screen.getByLabelText('Amount raised'), '750.50');
    await userEvent.type(screen.getByLabelText('Evidence URL'), 'https://example.com/seed');
    await userEvent.type(screen.getByLabelText('Source label'), 'Company site');
    await userEvent.click(screen.getByRole('button', { name: 'Add funding round' }));

    await waitFor(() => {
      expect(createAdminStartupFunding).toHaveBeenCalledWith('mercor', {
        roundLabel: 'Seed',
        announcementDate: '2024-01-02',
        amountRaisedMinor: 75050,
        valuationMinor: null,
        valuationBasis: 'unspecified',
        currency: 'USD',
        evidenceUrl: 'https://example.com/seed',
        sourceLabel: 'Company site',
        note: '',
      });
    });
    expect(await screen.findByText('Funding saved.')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Edit Seed' })).toBeTruthy();
  });

  it('edits an existing funding round', async () => {
    const rounds = [{
      id: 9,
      roundLabel: 'Seed',
      announcementDate: '2024-01-02',
      amountRaisedMinor: 10000,
      valuationMinor: null,
      valuationBasis: 'unspecified',
      currency: 'USD',
      evidenceUrl: 'https://example.com/seed',
      sourceLabel: 'Company site',
      note: '',
    }];
    getStartupFunding.mockResolvedValue({ rounds, latestReportedValuation: null });
    updateAdminStartupFunding.mockResolvedValue({ ...rounds[0], roundLabel: 'Seed extension' });

    renderFunding();
    await userEvent.click(await screen.findByRole('button', { name: 'Edit Seed' }));
    const label = screen.getByLabelText('Round label');
    await userEvent.clear(label);
    await userEvent.type(label, 'Seed extension');
    await userEvent.selectOptions(screen.getByLabelText('Valuation basis'), 'pre-money');
    await userEvent.type(screen.getByLabelText('Reported valuation'), '20');
    await userEvent.click(screen.getByRole('button', { name: 'Save funding correction' }));

    await waitFor(() => {
      expect(updateAdminStartupFunding).toHaveBeenCalledWith('mercor', 9, expect.objectContaining({
        roundLabel: 'Seed extension',
        amountRaisedMinor: 10000,
        valuationMinor: 2000,
        valuationBasis: 'pre-money',
      }));
    });
    expect(await screen.findByText('Funding saved.')).toBeTruthy();
  });

  it('shows validation and save errors', async () => {
    getStartupFunding.mockResolvedValue({ rounds: [], latestReportedValuation: null });
    renderFunding();
    await screen.findByRole('button', { name: 'Add funding round' });

    await userEvent.click(screen.getByRole('button', { name: 'Add funding round' }));
    expect(screen.getByRole('alert').textContent).toContain('round label');
    expect(createAdminStartupFunding).not.toHaveBeenCalled();

    await userEvent.type(screen.getByLabelText('Round label'), 'Seed');
    await userEvent.type(screen.getByLabelText('Announcement date'), '2024-01-02');
    await userEvent.type(screen.getByLabelText('Amount raised'), '0');
    await userEvent.type(screen.getByLabelText('Evidence URL'), 'https://example.com/seed');
    await userEvent.type(screen.getByLabelText('Source label'), 'Company site');
    await userEvent.click(screen.getByRole('button', { name: 'Add funding round' }));
    expect(screen.getByRole('alert').textContent).toContain('positive amount');

    await userEvent.clear(screen.getByLabelText('Amount raised'));
    createAdminStartupFunding.mockRejectedValueOnce(Object.assign(new Error('Check the startup fields and try again.'), { status: 400 }));
    await userEvent.click(screen.getByRole('button', { name: 'Add funding round' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Check the startup fields');
  });
});