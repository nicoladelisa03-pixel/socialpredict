import React, { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import AdminStartupFunding from './AdminStartupFunding';
import {
  createAdminStartup,
  listAdminStartupMarkets,
  listStartups,
  setAdminMarketStartup,
  updateAdminStartup,
} from '../../../api/startupsApi';

const slugPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

const fieldClass = 'w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40';
const labelClass = 'block text-xs font-mono uppercase tracking-[0.16em] text-gray-400';
const buttonClass = 'rounded-md bg-primary-pink px-4 py-2 text-sm font-semibold text-white transition hover:bg-pink-500 disabled:cursor-not-allowed disabled:bg-gray-600';

const emptyCreate = {
  name: '',
  slug: '',
  description: '',
  websiteUrl: '',
};

const validateStartup = ({ name, slug, description, websiteUrl, requireSlug }) => {
  const errors = [];
  const trimmedName = String(name || '').trim();
  const trimmedSlug = String(slug || '').trim();
  const trimmedDescription = String(description || '').trim();
  const website = String(websiteUrl || '').trim();

  if (!trimmedName) {
    errors.push('Enter a startup name.');
  } else if (trimmedName.length > 160) {
    errors.push('Name must be 160 characters or fewer.');
  }
  if (requireSlug) {
    if (!slugPattern.test(trimmedSlug) || trimmedSlug.length > 64) {
      errors.push('Use a slug of lowercase letters, numbers, and single hyphens.');
    }
  }
  if (trimmedDescription.length > 2000) {
    errors.push('Description must be 2000 characters or fewer.');
  }
  if (website.length > 500) {
    errors.push('Website must be 500 characters or fewer.');
  } else if (website) {
    try {
      const url = new URL(website);
      if (url.protocol !== 'http:' && url.protocol !== 'https:') {
        errors.push('Website must start with http:// or https://.');
      }
    } catch {
      errors.push('Enter a valid website URL, or leave it blank.');
    }
  }
  return errors;
};

const requestMessage = (error, fallback) => {
  if (error?.status === 401) {
    return 'Sign in again to manage startups.';
  }
  if (error?.status === 403) {
    return 'You need an admin account to manage startups.';
  }
  if (error?.status === 409) {
    return error.message || 'A startup with this slug already exists.';
  }
  return error?.message || fallback;
};

const startupLabel = (startups, slug) => {
  if (!slug) {
    return 'not linked';
  }
  const startup = startups.find((item) => item.slug === slug);
  return startup ? `${startup.name} (${startup.slug})` : slug;
};

function AdminStartups() {
  const [startups, setStartups] = useState([]);
  const [markets, setMarkets] = useState([]);
  const [status, setStatus] = useState('loading');
  const [loadError, setLoadError] = useState('');
  const [requestId, setRequestId] = useState(0);
  const [createForm, setCreateForm] = useState(emptyCreate);
  const [createError, setCreateError] = useState('');
  const [creating, setCreating] = useState(false);
  const [selectedSlug, setSelectedSlug] = useState('');
  const [editForm, setEditForm] = useState({ name: '', description: '', websiteUrl: '' });
  const [editError, setEditError] = useState('');
  const [editMessage, setEditMessage] = useState('');
  const [savingEdit, setSavingEdit] = useState(false);
  const [marketId, setMarketId] = useState('');
  const [linkSlug, setLinkSlug] = useState('');
  const [linkError, setLinkError] = useState('');
  const [linkMessage, setLinkMessage] = useState('');
  const [linking, setLinking] = useState(false);
  const hasLoaded = useRef(false);

  useEffect(() => {
    let ignore = false;
    const load = async () => {
      if (!hasLoaded.current) {
        setStatus('loading');
      }
      setLoadError('');
      try {
        const [startupResult, marketResult] = await Promise.all([
          listStartups(),
          listAdminStartupMarkets(),
        ]);
        if (ignore) {
          return;
        }
        setStartups(Array.isArray(startupResult?.startups) ? startupResult.startups : []);
        setMarkets(Array.isArray(marketResult?.markets) ? marketResult.markets : []);
        hasLoaded.current = true;
        setStatus('ready');
      } catch (error) {
        if (ignore) {
          return;
        }
        setStartups([]);
        setMarkets([]);
        setLoadError(requestMessage(error, 'Unable to load startups.'));
        setStatus('error');
      }
    };
    load();
    return () => {
      ignore = true;
    };
  }, [requestId]);

  const selectedStartup = startups.find((startup) => startup.slug === selectedSlug) || null;
  const selectedMarket = markets.find((market) => String(market.marketId) === String(marketId)) || null;
  const linkedMarkets = selectedStartup
    ? markets.filter((market) => market.startupSlug === selectedStartup.slug)
    : [];

  const chooseStartup = (startup) => {
    setSelectedSlug(startup.slug);
    setEditForm({
      name: startup.name || '',
      description: startup.description || '',
      websiteUrl: startup.websiteUrl || '',
    });
    setEditError('');
    setEditMessage('');
  };

  const chooseMarket = (nextMarketId) => {
    const market = markets.find((item) => String(item.marketId) === String(nextMarketId));
    setMarketId(nextMarketId);
    setLinkSlug(market?.startupSlug || '');
    setLinkError('');
    setLinkMessage('');
  };

  const handleCreate = async (event) => {
    event.preventDefault();
    if (creating) {
      return;
    }
    const errors = validateStartup({ ...createForm, requireSlug: true });
    if (errors.length > 0) {
      setCreateError(errors[0]);
      return;
    }
    setCreating(true);
    setCreateError('');
    try {
      const created = await createAdminStartup({
        name: createForm.name.trim(),
        slug: createForm.slug.trim(),
        description: createForm.description.trim(),
        websiteUrl: createForm.websiteUrl.trim(),
      });
      setCreateForm(emptyCreate);
      setRequestId((current) => current + 1);
      if (created?.slug) {
        setSelectedSlug(created.slug);
        setEditForm({
          name: created.name || createForm.name.trim(),
          description: created.description || '',
          websiteUrl: created.websiteUrl || '',
        });
        setEditMessage('Startup created.');
      }
    } catch (error) {
      setCreateError(requestMessage(error, 'Unable to create this startup.'));
    } finally {
      setCreating(false);
    }
  };

  const handleEdit = async (event) => {
    event.preventDefault();
    if (savingEdit || !selectedStartup) {
      return;
    }
    const errors = validateStartup({ ...editForm, requireSlug: false });
    if (errors.length > 0) {
      setEditError(errors[0]);
      setEditMessage('');
      return;
    }
    setSavingEdit(true);
    setEditError('');
    setEditMessage('');
    try {
      const updated = await updateAdminStartup(selectedStartup.slug, {
        name: editForm.name.trim(),
        description: editForm.description.trim(),
        websiteUrl: editForm.websiteUrl.trim(),
      });
      setEditMessage('Startup saved.');
      setEditForm({
        name: updated?.name || editForm.name.trim(),
        description: updated?.description || '',
        websiteUrl: updated?.websiteUrl || '',
      });
      setRequestId((current) => current + 1);
    } catch (error) {
      setEditError(requestMessage(error, 'Unable to update this startup.'));
    } finally {
      setSavingEdit(false);
    }
  };

  const handleLink = async (event) => {
    event.preventDefault();
    if (linking || !selectedMarket) {
      return;
    }
    setLinking(true);
    setLinkError('');
    setLinkMessage('');
    try {
      const saved = await setAdminMarketStartup(selectedMarket.marketId, linkSlug);
      const savedSlug = saved?.startupSlug || '';
      setLinkSlug(savedSlug);
      setLinkMessage(savedSlug
        ? `Saved. This market is linked to ${startupLabel(startups, savedSlug)}.`
        : 'Saved. This market is not linked to a startup.');
      setRequestId((current) => current + 1);
    } catch (error) {
      setLinkError(requestMessage(error, 'Unable to update this market\'s startup.'));
    } finally {
      setLinking(false);
    }
  };

  return (
    <div className='grid gap-6 text-white'>
      <div>
        <h1 className='text-2xl font-bold'>Startups</h1>
        <p className='mt-1 text-sm text-gray-300'>
          Create a startup, update its public details, and choose which existing markets belong to it.
        </p>
      </div>

      {status === 'loading' && (
        <p className='text-sm text-gray-300' role='status'>Loading startups...</p>
      )}

      {status === 'error' && (
        <div className='rounded-lg border border-red-600 bg-red-900/50 p-4' role='alert'>
          <p className='mb-3 text-red-200'>{loadError}</p>
          <button type='button' className={buttonClass} onClick={() => setRequestId((current) => current + 1)}>
            Try again
          </button>
        </div>
      )}

      {status === 'ready' && (
        <>
          <form onSubmit={handleCreate} className='grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4'>
            <h2 className='text-lg font-semibold'>Create a startup</h2>
            <p className='text-sm text-gray-300'>The slug cannot be changed after creation.</p>
            <label className={labelClass} htmlFor='startup-name'>
              Name
              <input
                id='startup-name'
                className={`${fieldClass} mt-2 normal-case tracking-normal`}
                value={createForm.name}
                onChange={(event) => setCreateForm((current) => ({ ...current, name: event.target.value }))}
                maxLength={160}
                required
              />
            </label>
            <label className={labelClass} htmlFor='startup-slug'>
              Slug
              <input
                id='startup-slug'
                className={`${fieldClass} mt-2 normal-case tracking-normal`}
                value={createForm.slug}
                onChange={(event) => setCreateForm((current) => ({ ...current, slug: event.target.value.toLowerCase() }))}
                maxLength={64}
                required
                spellCheck='false'
              />
            </label>
            <label className={labelClass} htmlFor='startup-description'>
              Description
              <textarea
                id='startup-description'
                className={`${fieldClass} mt-2 min-h-[6rem] normal-case tracking-normal`}
                value={createForm.description}
                onChange={(event) => setCreateForm((current) => ({ ...current, description: event.target.value }))}
                maxLength={2000}
              />
            </label>
            <label className={labelClass} htmlFor='startup-website'>
              Website
              <input
                id='startup-website'
                className={`${fieldClass} mt-2 normal-case tracking-normal`}
                value={createForm.websiteUrl}
                onChange={(event) => setCreateForm((current) => ({ ...current, websiteUrl: event.target.value }))}
                maxLength={500}
                placeholder='https://'
              />
            </label>
            {createError && <p className='text-sm text-red-300' role='alert'>{createError}</p>}
            <button type='submit' className={buttonClass} disabled={creating}>
              {creating ? 'Creating...' : 'Create startup'}
            </button>
          </form>

          <section className='grid gap-3'>
            <h2 className='text-lg font-semibold'>Existing startups</h2>
            {startups.length === 0 ? (
              <p className='rounded-lg border border-gray-700 bg-gray-900/70 p-4 text-sm text-gray-300' role='status'>
                No startups yet. Create one to link markets to it.
              </p>
            ) : (
              <ul className='grid gap-2'>
                {startups.map((startup) => (
                  <li key={startup.slug}>
                    <button
                      type='button'
                      className={`w-full rounded-lg border p-3 text-left ${selectedSlug === startup.slug ? 'border-primary-pink bg-gray-800' : 'border-gray-700 bg-gray-900/70'}`}
                      onClick={() => chooseStartup(startup)}
                    >
                      <span className='block font-semibold'>{startup.name}</span>
                      <span className='block text-sm text-gray-400'>{startup.slug}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </section>

          {selectedStartup && (
            <form onSubmit={handleEdit} className='grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4'>
              <h2 className='text-lg font-semibold'>Edit {selectedStartup.name}</h2>
              <p className='text-sm text-gray-300'>
                Slug: <span className='font-mono text-white'>{selectedStartup.slug}</span>. The slug cannot be changed after creation.
              </p>
              <label className={labelClass} htmlFor='edit-startup-name'>
                Name
                <input
                  id='edit-startup-name'
                  className={`${fieldClass} mt-2 normal-case tracking-normal`}
                  value={editForm.name}
                  onChange={(event) => setEditForm((current) => ({ ...current, name: event.target.value }))}
                  maxLength={160}
                  required
                />
              </label>
              <label className={labelClass} htmlFor='edit-startup-description'>
                Description
                <textarea
                  id='edit-startup-description'
                  className={`${fieldClass} mt-2 min-h-[6rem] normal-case tracking-normal`}
                  value={editForm.description}
                  onChange={(event) => setEditForm((current) => ({ ...current, description: event.target.value }))}
                  maxLength={2000}
                />
              </label>
              <label className={labelClass} htmlFor='edit-startup-website'>
                Website
                <input
                  id='edit-startup-website'
                  className={`${fieldClass} mt-2 normal-case tracking-normal`}
                  value={editForm.websiteUrl}
                  onChange={(event) => setEditForm((current) => ({ ...current, websiteUrl: event.target.value }))}
                  maxLength={500}
                  placeholder='https://'
                />
              </label>
              {editError && <p className='text-sm text-red-300' role='alert'>{editError}</p>}
              {editMessage && <p className='text-sm text-green-300' role='status'>{editMessage}</p>}
              <div className='flex flex-wrap gap-3'>
                <button type='submit' className={buttonClass} disabled={savingEdit}>
                  {savingEdit ? 'Saving...' : 'Save startup'}
                </button>
                <Link className='self-center text-sm text-blue-400 hover:text-blue-300' to={`/startups/${selectedStartup.slug}`}>
                  View public page
                </Link>
              </div>
              <div>
                <h3 className='mb-2 text-sm font-semibold'>Markets linked to this startup</h3>
                {linkedMarkets.length === 0 ? (
                  <p className='text-sm text-gray-300' role='status'>No markets are linked to this startup.</p>
                ) : (
                  <ul className='grid gap-2'>
                    {linkedMarkets.map((market) => (
                      <li key={market.marketId} className='flex flex-wrap items-center justify-between gap-2 rounded-md border border-gray-700 px-3 py-2'>
                        <span className='text-sm'>
                          #{market.marketId} {market.questionTitle}
                          <span className='ml-2 text-gray-400'>{market.lifecycleStatus || 'published'}</span>
                        </span>
                        <button type='button' className='text-sm text-blue-400 hover:text-blue-300' onClick={() => chooseMarket(String(market.marketId))}>
                          Change link
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
                <p className='mt-2 text-xs text-gray-400'>
                  Proposed markets stay off the public startup page until they are approved.
                </p>
              </div>
            </form>
          )}

          {selectedStartup ? <AdminStartupFunding slug={selectedStartup.slug} /> : null}

          <form onSubmit={handleLink} className='grid gap-4 rounded-lg border border-gray-700 bg-gray-900/70 p-4'>
            <h2 className='text-lg font-semibold'>Link a market</h2>
            {markets.length === 0 ? (
              <p className='text-sm text-gray-300' role='status'>No markets are available to link.</p>
            ) : (
              <>
                <label className={labelClass} htmlFor='market-link-market'>
                  Market
                  <select
                    id='market-link-market'
                    className={`${fieldClass} mt-2 normal-case tracking-normal`}
                    value={marketId}
                    onChange={(event) => chooseMarket(event.target.value)}
                  >
                    <option value=''>Choose a market</option>
                    {markets.map((market) => (
                      <option key={market.marketId} value={String(market.marketId)}>
                        #{market.marketId} {market.questionTitle}
                      </option>
                    ))}
                  </select>
                </label>
                {selectedMarket && (
                  <p className='text-sm text-gray-300' role='status'>
                    Current link: {startupLabel(startups, selectedMarket.startupSlug)}
                  </p>
                )}
                <label className={labelClass} htmlFor='market-link-startup'>
                  Startup
                  <select
                    id='market-link-startup'
                    className={`${fieldClass} mt-2 normal-case tracking-normal`}
                    value={linkSlug}
                    onChange={(event) => setLinkSlug(event.target.value)}
                    disabled={!selectedMarket}
                  >
                    <option value=''>No startup</option>
                    {startups.map((startup) => (
                      <option key={startup.slug} value={startup.slug}>
                        {startup.name}
                      </option>
                    ))}
                  </select>
                </label>
                {linkError && <p className='text-sm text-red-300' role='alert'>{linkError}</p>}
                {linkMessage && <p className='text-sm text-green-300' role='status'>{linkMessage}</p>}
                <button type='submit' className={buttonClass} disabled={linking || !selectedMarket}>
                  {linking ? 'Saving...' : 'Save market link'}
                </button>
              </>
            )}
          </form>
        </>
      )}
    </div>
  );
}

export default AdminStartups;
