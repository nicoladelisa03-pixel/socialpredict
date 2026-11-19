import React, { useEffect, useState } from 'react';
import { getStartupFunding } from '../../api/startupsApi';
import { formatMinorAmount, reportedValuationDisclaimer, valuationBasisText } from './fundingFormat';

function moneyLine(label, minor, currency) {
  if (minor == null) {
    return `${label} was not disclosed`;
  }
  return `${label} ${formatMinorAmount(minor, currency)}`;
}

function StartupFunding({ slug }) {
  const [history, setHistory] = useState(null);
  const [status, setStatus] = useState('loading');
  const [errorMessage, setErrorMessage] = useState('');
  const [requestId, setRequestId] = useState(0);

  useEffect(() => {
    let ignore = false;
    const load = async () => {
      setStatus('loading');
      setErrorMessage('');
      try {
        const result = await getStartupFunding(slug);
        if (ignore) {
          return;
        }
        setHistory(result || { rounds: [], latestReportedValuation: null });
        setStatus('ready');
      } catch (error) {
        if (ignore) {
          return;
        }
        setHistory(null);
        setErrorMessage(error?.message || 'Unable to load funding history.');
        setStatus('error');
      }
    };
    load();
    return () => {
      ignore = true;
    };
  }, [slug, requestId]);

  const rounds = Array.isArray(history?.rounds) ? history.rounds : [];
  const latest = history?.latestReportedValuation || null;

  return (
    <section className='mt-6' aria-label='Funding history'>
      <h2 className='text-xl font-semibold mb-3'>Funding history</h2>
      {status === 'loading' && <p role='status'>Loading funding history...</p>}
      {status === 'error' && (
        <div className='rounded-lg border border-red-600 bg-red-900/50 p-4' role='alert'>
          <p className='text-red-300 mb-3'>{errorMessage}</p>
          <button
            type='button'
            onClick={() => setRequestId((current) => current + 1)}
            className='px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors'
          >
            Retry funding
          </button>
        </div>
      )}
      {status === 'ready' && rounds.length === 0 && (
        <p className='rounded-lg bg-gray-800 p-4 text-gray-300'>No verified funding data added yet</p>
      )}
      {status === 'ready' && latest && (
        <div className='mb-4 rounded-lg border border-gray-600 bg-gray-800 p-4'>
          <h3 className='text-lg font-semibold'>Latest reported valuation</h3>
          <p className='mt-2 text-gray-100 break-words'>
            {formatMinorAmount(latest.valuationMinor, latest.currency)}
          </p>
          <p className='mt-1 text-sm text-gray-300'>
            Reported on {latest.announcementDate}. {valuationBasisText(latest.valuationBasis)}.
          </p>
          <p className='mt-2 text-sm text-gray-300'>{reportedValuationDisclaimer}</p>
          <a
            href={latest.evidenceUrl}
            target='_blank'
            rel='noopener noreferrer'
            className='mt-2 inline-block text-blue-400 hover:text-blue-300 break-all'
          >
            Source: {latest.sourceLabel}
          </a>
        </div>
      )}
      {status === 'ready' && rounds.length > 0 && (
        <ol className='grid gap-3'>
          {rounds.map((round) => (
            <li key={round.id} className='rounded-lg bg-gray-800 p-4'>
              <h3 className='text-lg font-semibold break-words'>{round.roundLabel}</h3>
              <p className='mt-1 text-sm text-gray-400'>{round.announcementDate}</p>
              <p className='mt-2 text-gray-200 break-words'>{moneyLine('Amount raised', round.amountRaisedMinor, round.currency)}</p>
              <p className='text-gray-200 break-words'>
                {round.valuationMinor == null
                  ? 'Reported valuation was not disclosed'
                  : `Reported valuation ${formatMinorAmount(round.valuationMinor, round.currency)} (${valuationBasisText(round.valuationBasis)})`}
              </p>
              {round.note ? <p className='mt-2 text-sm text-gray-300 break-words'>{round.note}</p> : null}
              <a
                href={round.evidenceUrl}
                target='_blank'
                rel='noopener noreferrer'
                className='mt-2 inline-block text-blue-400 hover:text-blue-300 break-all'
              >
                {round.sourceLabel}
              </a>
            </li>
          ))}
        </ol>
      )}
    </section>
  );
}

export default StartupFunding;
