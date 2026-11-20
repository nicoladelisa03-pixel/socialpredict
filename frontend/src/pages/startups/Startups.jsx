import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { listStartups } from '../../api/startupsApi';
import { SITE_NAME } from '../../config';
import { useDocumentMeta } from '../../hooks/useDocumentMeta';
import { isTestStartup, safeStartupWebsite, startupWebsiteLabel } from './startupWebsite';

function Startups() {
  useDocumentMeta({ title: `Startups | ${SITE_NAME}` });
  const [startups, setStartups] = useState([]);
  const [status, setStatus] = useState('loading');
  const [errorMessage, setErrorMessage] = useState('');
  const [requestId, setRequestId] = useState(0);

  useEffect(() => {
    let ignore = false;

    const load = async () => {
      setStatus('loading');
      setErrorMessage('');
      try {
        const result = await listStartups();
        if (ignore) {
          return;
        }
        const items = Array.isArray(result?.startups) ? result.startups : [];
        setStartups(items);
        setStatus('ready');
      } catch (error) {
        if (ignore) {
          return;
        }
        setStartups([]);
        setErrorMessage(error?.message || 'Unable to load startups.');
        setStatus('error');
      }
    };

    load();
    return () => {
      ignore = true;
    };
  }, [requestId]);

  return (
    <div className='flex flex-col text-white'>
      <h1 className='text-2xl md:text-3xl font-bold mb-2'>Startups</h1>
      <p className='text-gray-300 mb-4 md:mb-6'>
        Browse the startups saved for this prediction market.
      </p>

      {status === 'loading' && (
        <p className='text-lg' role='status'>Loading startups...</p>
      )}

      {status === 'error' && (
        <div className='rounded-lg border border-red-600 bg-red-900/50 p-4' role='alert'>
          <p className='text-red-300 mb-3'>{errorMessage}</p>
          <button
            type='button'
            onClick={() => setRequestId((current) => current + 1)}
            className='px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors'
          >
            Try again
          </button>
        </div>
      )}

      {status === 'ready' && startups.length === 0 && (
        <p className='rounded-lg bg-gray-800 p-4 text-gray-300'>
          No startups have been added yet.
        </p>
      )}

      {status === 'ready' && startups.length > 0 && (
        <ul className='grid grid-cols-1 gap-4 md:grid-cols-2'>
          {startups.map((startup) => {
            const name = startup.name || startup.slug;
            const website = safeStartupWebsite(startup.websiteUrl);
            const profilePath = startup.slug ? `/startups/${encodeURIComponent(startup.slug)}` : '';
            if (!name) {
              return null;
            }
            return (
              <li key={startup.slug || startup.id} className='rounded-lg bg-gray-800 shadow-lg overflow-hidden'>
                {isTestStartup(startup.slug) ? (
                  <p className='px-4 pt-4 text-xs font-semibold uppercase tracking-wide text-gray-400 md:px-6'>Test record</p>
                ) : null}
                {profilePath ? (
                  <Link
                    to={profilePath}
                    aria-label={name}
                    className='block p-4 md:p-6 hover:bg-gray-700/50 transition-colors duration-200'
                  >
                    <h2 className='text-xl font-semibold'>{name}</h2>
                    {startup.description ? (
                      <p className='mt-2 text-gray-300'>{startup.description}</p>
                    ) : null}
                  </Link>
                ) : (
                  <div className='p-4 md:p-6'>
                    <h2 className='text-xl font-semibold'>{name}</h2>
                  </div>
                )}
                {website ? (
                  <div className='px-4 pb-4 md:px-6 md:pb-6'>
                    <a
                      href={website}
                      target='_blank'
                      rel='noopener noreferrer'
                      className='inline-block text-blue-400 hover:text-blue-300 break-all transition-colors duration-200'
                    >
                      {startupWebsiteLabel(website)}
                    </a>
                  </div>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

export default Startups;
