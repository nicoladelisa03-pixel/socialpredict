import React, { useEffect, useState } from 'react';
import {
  createAdminStartupFunding,
  getStartupFunding,
  updateAdminStartupFunding,
} from '../../../api/startupsApi';
import { majorUnitsToMinor } from '../../../pages/startups/fundingFormat';

const fieldClass = 'w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40';
const labelClass = 'block text-xs font-mono uppercase tracking-[0.16em] text-gray-400';
const buttonClass = 'rounded-md bg-primary-pink px-4 py-2 text-sm font-semibold text-white transition hover:bg-pink-500 disabled:cursor-not-allowed disabled:bg-gray-600';

const emptyForm = {
  roundLabel: '',
  announcementDate: '',
  amountRaised: '',
  valuation: '',
  valuationBasis: 'unspecified',
  currency: 'USD',
  evidenceUrl: '',
  sourceLabel: '',
  note: '',
};

const requestMessage = (error, fallback) => {
  if (error?.status === 401) {
    return 'Sign in again to manage startups.';
  }
  if (error?.status === 403) {
    return 'You need an admin account to manage startups.';
  }
  return error?.message || fallback;
};

const minorToMajorInput = (minor) => {
  if (minor == null) {
    return '';
  }
  const value = Number(minor);
  if (!Number.isSafeInteger(value) || value <= 0) {
    return '';
  }
  const whole = Math.trunc(value / 100);
  const cents = String(value % 100).padStart(2, '0');
  return cents === '00' ? String(whole) : `${whole}.${cents}`;
};

function validateFunding(form) {
  const errors = [];
  if (!form.roundLabel.trim() || form.roundLabel.trim().length > 80) {
    errors.push('Enter a round label of 80 characters or fewer.');
  }
  if (!/^\d{4}-\d{2}-\d{2}$/.test(form.announcementDate.trim())) {
    errors.push('Enter an announcement date as YYYY-MM-DD.');
  }
  try {
    majorUnitsToMinor(form.amountRaised);
    majorUnitsToMinor(form.valuation);
  } catch (error) {
    errors.push(error.message);
  }
  if (!['pre-money', 'post-money', 'unspecified'].includes(form.valuationBasis)) {
    errors.push('Choose a valuation basis.');
  }
  if (!/^[A-Za-z]{3}$/.test(form.currency.trim())) {
    errors.push('Enter a three-letter currency code.');
  }
  const evidence = form.evidenceUrl.trim();
  if (!/^https?:\/\//i.test(evidence) || evidence.length > 500) {
    errors.push('Enter an http or https evidence link.');
  }
  if (!form.sourceLabel.trim() || form.sourceLabel.trim().length > 120) {
    errors.push('Enter a short source label.');
  }
  if (form.note.trim().length > 500) {
    errors.push('Keep the note to 500 characters or fewer.');
  }
  return errors;
}

function formToPayload(form) {
  return {
    roundLabel: form.roundLabel.trim(),
    announcementDate: form.announcementDate.trim(),
    amountRaisedMinor: majorUnitsToMinor(form.amountRaised),
    valuationMinor: majorUnitsToMinor(form.valuation),
    valuationBasis: form.valuation.trim() ? form.valuationBasis : 'unspecified',
    currency: form.currency.trim().toUpperCase(),
    evidenceUrl: form.evidenceUrl.trim(),
    sourceLabel: form.sourceLabel.trim(),
    note: form.note.trim(),
  };
}

function AdminStartupFunding({ slug }) {
  const [rounds, setRounds] = useState([]);
  const [status, setStatus] = useState('loading');
  const [loadError, setLoadError] = useState('');
  const [requestId, setRequestId] = useState(0);
  const [form, setForm] = useState(emptyForm);
  const [editingId, setEditingId] = useState(null);
  const [formError, setFormError] = useState('');
  const [saveMessage, setSaveMessage] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let ignore = false;
    const load = async () => {
      setStatus('loading');
      setLoadError('');
      try {
        const result = await getStartupFunding(slug);
        if (ignore) {
          return;
        }
        setRounds(Array.isArray(result?.rounds) ? result.rounds : []);
        setStatus('ready');
      } catch (error) {
        if (ignore) {
          return;
        }
        setRounds([]);
        setLoadError(requestMessage(error, 'Unable to load funding history.'));
        setStatus('error');
      }
    };
    load();
    return () => {
      ignore = true;
    };
  }, [slug, requestId]);

  const updateField = (field) => (event) => {
    setForm((current) => ({ ...current, [field]: event.target.value }));
  };

  const beginEdit = (round) => {
    setEditingId(round.id);
    setFormError('');
    setSaveMessage('');
    setForm({
      roundLabel: round.roundLabel || '',
      announcementDate: round.announcementDate || '',
      amountRaised: minorToMajorInput(round.amountRaisedMinor),
      valuation: minorToMajorInput(round.valuationMinor),
      valuationBasis: round.valuationBasis || 'unspecified',
      currency: round.currency || 'USD',
      evidenceUrl: round.evidenceUrl || '',
      sourceLabel: round.sourceLabel || '',
      note: round.note || '',
    });
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (saving) {
      return;
    }
    const errors = validateFunding(form);
    if (errors.length > 0) {
      setFormError(errors[0]);
      setSaveMessage('');
      return;
    }
    setSaving(true);
    setFormError('');
    setSaveMessage('');
    try {
      const payload = formToPayload(form);
      if (editingId) {
        await updateAdminStartupFunding(slug, editingId, payload);
      } else {
        await createAdminStartupFunding(slug, payload);
      }
      setForm(emptyForm);
      setEditingId(null);
      setSaveMessage('Funding saved.');
      setRequestId((current) => current + 1);
    } catch (error) {
      setFormError(requestMessage(error, 'Unable to save this funding round.'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className='grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4' aria-label='Funding records'>
      <div>
        <h2 className='text-lg font-semibold'>Funding records</h2>
        <p className='mt-1 text-sm text-gray-300'>
          Leave an amount blank when it was not disclosed. Zero is not a stand-in for unknown.
        </p>
      </div>
      {status === 'loading' && <p role='status'>Loading funding records...</p>}
      {status === 'error' && (
        <div role='alert' className='rounded-md border border-red-600 bg-red-900/40 p-3 text-sm text-red-200'>
          {loadError}
        </div>
      )}
      {status === 'ready' && rounds.length === 0 && (
        <p className='text-sm text-gray-300' role='status'>No verified funding data added yet</p>
      )}
      {status === 'ready' && rounds.length > 0 && (
        <ul className='grid gap-2'>
          {rounds.map((round) => (
            <li key={round.id} className='flex flex-wrap items-center justify-between gap-2 rounded-md border border-gray-700 px-3 py-2'>
              <span className='text-sm break-words'>{round.roundLabel} · {round.announcementDate}</span>
              <button type='button' className='text-sm text-blue-400 hover:text-blue-300' onClick={() => beginEdit(round)}>
                Edit {round.roundLabel}
              </button>
            </li>
          ))}
        </ul>
      )}
      <form onSubmit={handleSubmit} className='grid gap-3'>
        <label className={labelClass} htmlFor='funding-round-label'>
          Round label
          <input id='funding-round-label' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.roundLabel} onChange={updateField('roundLabel')} maxLength={80} />
        </label>
        <label className={labelClass} htmlFor='funding-date'>
          Announcement date
          <input id='funding-date' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.announcementDate} onChange={updateField('announcementDate')} placeholder='YYYY-MM-DD' />
        </label>
        <label className={labelClass} htmlFor='funding-amount'>
          Amount raised
          <input id='funding-amount' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.amountRaised} onChange={updateField('amountRaised')} inputMode='decimal' placeholder='Blank if unknown' />
        </label>
        <label className={labelClass} htmlFor='funding-valuation'>
          Reported valuation
          <input id='funding-valuation' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.valuation} onChange={updateField('valuation')} inputMode='decimal' placeholder='Blank if unknown' />
        </label>
        <label className={labelClass} htmlFor='funding-basis'>
          Valuation basis
          <select id='funding-basis' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.valuationBasis} onChange={updateField('valuationBasis')}>
            <option value='unspecified'>Unspecified</option>
            <option value='pre-money'>Pre-money</option>
            <option value='post-money'>Post-money</option>
          </select>
        </label>
        <label className={labelClass} htmlFor='funding-currency'>
          Currency
          <input id='funding-currency' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.currency} onChange={updateField('currency')} maxLength={3} />
        </label>
        <label className={labelClass} htmlFor='funding-evidence'>
          Evidence URL
          <input id='funding-evidence' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.evidenceUrl} onChange={updateField('evidenceUrl')} placeholder='https://' />
        </label>
        <label className={labelClass} htmlFor='funding-source'>
          Source label
          <input id='funding-source' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.sourceLabel} onChange={updateField('sourceLabel')} maxLength={120} />
        </label>
        <label className={labelClass} htmlFor='funding-note'>
          Note
          <input id='funding-note' className={`${fieldClass} mt-2 normal-case tracking-normal`} value={form.note} onChange={updateField('note')} maxLength={500} />
        </label>
        {formError && <p className='text-sm text-red-300' role='alert'>{formError}</p>}
        {saveMessage && <p className='text-sm text-green-300' role='status'>{saveMessage}</p>}
        <div className='flex flex-wrap gap-3'>
          <button type='submit' className={buttonClass} disabled={saving}>
            {saving ? 'Saving...' : (editingId ? 'Save funding correction' : 'Add funding round')}
          </button>
          {editingId ? (
            <button
              type='button'
              className='text-sm text-blue-400 hover:text-blue-300'
              onClick={() => {
                setEditingId(null);
                setForm(emptyForm);
                setFormError('');
              }}
            >
              Cancel edit
            </button>
          ) : null}
        </div>
      </form>
    </section>
  );
}

export default AdminStartupFunding;
