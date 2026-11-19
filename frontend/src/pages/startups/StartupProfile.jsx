import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getStartup, getStartupMarkets } from '../../api/startupsApi';
import { safeStartupWebsite, startupWebsiteLabel } from './startupWebsite';

const missingDescription = 'No description has been added yet.';

function StartupProfile() {
  const { slug } = useParams();
  const [startup, setStartup] = useState(null);
  const [status, setStatus] = useState('loading');
  const [errorMessage, setErrorMessage] = useState('');
  const [requestId, setRequestId] = useState(0);
  const [markets, setMarkets] = useState([]);
  const [marketsStatus, setMarketsStatus] = useState('loading');
  const [marketsError, setMarketsError] = useState('');
  const [marketsRequestId, setMarketsRequestId] = useState(0);

  useEffect(() => {
    let ignore = false;

    const load = async () => {
      setStatus('loading');
      setErrorMessage('');
      setStartup(null);
      try {
        const result = await getStartup(slug);
        if (ignore) {
          return;
        }
        setStartup(result || null);
        setStatus('ready');
      } catch (error) {
        if (ignore) {
          return;
        }
        setStartup(null);
        if (error?.status === 404) {
          setStatus('not-found');
          return;
        }
        setErrorMessage(error?.message || 'Unable to load this startup.');
        setStatus('error');
      }
    };

    load();
    return () => {
      ignore = true;
    };
  }, [slug, requestId]);

  useEffect(() => {
    let ignore = false;

    const loadMarkets = async () => {
      setMarketsStatus('loading');
      setMarketsError('');
      try {
        const result = await getStartupMarkets(slug);
        if (ignore) {
          return;
        }
        setMarkets(Array.isArray(result?.markets) ? result.markets : []);
        setMarketsStatus('ready');
      } catch (error) {
        if (ignore) {
          return;
        }
        setMarkets([]);
        setMarketsError(error?.message || 'Unable to load this startup\'s markets.');
        setMarketsStatus('error');
      }
    };

    loadMarkets();
    return () => {
      ignore = true;
    };
  }, [slug, marketsRequestId]);

  const name = startup?.name || startup?.slug || slug;
  const website = safeStartupWebsite(startup?.websiteUrl);

  return (
    <div className='flex flex-col text-white'>
      <Link to='/startups' className='mb-4 inline-block text-blue-400 hover:text-blue-300'>
        All startups
      </Link>

      {status === 'loading' && (
        <p className='text-lg' role='status'>Loading startup...</p>
      )}

      {status === 'not-found' && (
        <div className='rounded-lg border border-gray-600 bg-gray-800 p-4 md:p-6' role='status'>
          <h1 className='text-2xl font-bold mb-2'>Startup not found</h1>
          <p className='text-gray-300'>No startup matches this address.</p>
        </div>
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

      {status === 'ready' && (
        <article className='rounded-lg bg-gray-800 shadow-lg p-4 md:p-6'>
          <h1 className='text-2xl md:text-3xl font-bold'>{name}</h1>
          <p className='mt-3 text-gray-300'>{startup?.description || missingDescription}</p>
          {website ? (
            <a
              href={website}
              target='_blank'
              rel='noopener noreferrer'
              className='mt-4 inline-block text-blue-400 hover:text-blue-300 break-all transition-colors duration-200'
            >
              {startupWebsiteLabel(website)}
            </a>
          ) : (
            <p className='mt-4 text-gray-400'>No website has been added yet.</p>
          )}
        </article>
      )}

      {status === 'ready' && (
        <section className='mt-6' aria-label='Prediction markets'>
          <h2 className='text-xl font-semibold mb-3'>Prediction markets</h2>
          {marketsStatus === 'loading' && (
            <p role='status'>Loading markets...</p>
          )}
          {marketsStatus === 'error' && (
            <div className='rounded-lg border border-red-600 bg-red-900/50 p-4' role='alert'>
              <p className='text-red-300 mb-3'>{marketsError}</p>
              <button
                type='button'
                onClick={() => setMarketsRequestId((current) => current + 1)}
                className='px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors'
              >
                Try again
              </button>
            </div>
          )}
          {marketsStatus === 'ready' && markets.length === 0 && (
            <p className='rounded-lg bg-gray-800 p-4 text-gray-300'>No prediction markets are linked yet.</p>
          )}
          {marketsStatus === 'ready' && markets.length > 0 && (
            <ul className='grid grid-cols-1 gap-3'>
              {markets.map((market) => (
                <li key={market.id} className='rounded-lg bg-gray-800 p-4'>
                  <Link to={`/markets/${market.id}`} className='block text-lg font-semibold text-blue-400 hover:text-blue-300 break-words'>
                    {market.questionTitle}
                  </Link>
                  {market.isResolved ? (
                    <p className='mt-1 text-sm text-gray-400'>Resolved {market.resolutionResult}</p>
                  ) : null}
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </div>
  );
}

export default StartupProfile;
