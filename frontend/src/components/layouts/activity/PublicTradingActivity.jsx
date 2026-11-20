import React, { useEffect, useState } from 'react';
import { API_URL } from '../../../config';
import { unwrapApiResponse } from '../../../utils/apiResponse';

const numberValue = (value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
};

export const PublicTradingSummary = ({ activity, loading, error }) => {
  if (loading) {
    return <div className="p-4 text-sm text-gray-400">Loading market activity...</div>;
  }
  if (error) {
    return <div className="m-4 rounded-md bg-red-700 p-3 text-sm text-white">{error}</div>;
  }

  const betCount = numberValue(activity?.betCount);
  const yesVolume = numberValue(activity?.yesVolume);
  const noVolume = numberValue(activity?.noVolume);
  const participantCount = numberValue(activity?.participantCount);
  const empty = betCount === 0 && participantCount === 0;

  return (
    <div className="p-4 text-sm text-gray-200">
      <p className="mb-3 text-gray-400">
        Names, individual bets, share counts, and profit stay private. This view shows market totals only.
      </p>
      {empty ? (
        <p className="text-gray-300">No public trades yet.</p>
      ) : (
        <dl className="grid grid-cols-2 gap-3 sm:grid-cols-4">
          <div>
            <dt className="text-xs uppercase tracking-wide text-gray-400">Trades</dt>
            <dd className="text-lg text-white">{betCount}</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-gray-400">Participants</dt>
            <dd className="text-lg text-white">{participantCount}</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-gray-400">YES volume</dt>
            <dd className="text-lg text-white">{yesVolume}</dd>
          </div>
          <div>
            <dt className="text-xs uppercase tracking-wide text-gray-400">NO volume</dt>
            <dd className="text-lg text-white">{noVolume}</dd>
          </div>
        </dl>
      )}
    </div>
  );
};

const usePublicActivity = (path, refreshTrigger) => {
  const [activity, setActivity] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!path) {
      setActivity(null);
      setError('');
      return undefined;
    }

    let ignore = false;
    const load = async () => {
      setLoading(true);
      setError('');
      try {
        const response = await fetch(`${API_URL}${path}`);
        const payload = await response.json().catch(() => ({}));
        if (!response.ok) {
          throw new Error(payload.reason || payload.message || 'Failed to load market activity');
        }
        if (!ignore) {
          setActivity(unwrapApiResponse(payload));
        }
      } catch (err) {
        if (!ignore) {
          setActivity(null);
          setError(err.message || 'Failed to load market activity');
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };

    load();
    return () => {
      ignore = true;
    };
  }, [path, refreshTrigger]);

  return { activity, loading, error };
};

export const PublicMarketActivity = ({ marketId, refreshTrigger }) => {
  const state = usePublicActivity(marketId ? `/v0/markets/bets/${marketId}` : '', refreshTrigger);
  return <PublicTradingSummary {...state} />;
};

export const PublicGroupActivity = ({ groupId, refreshTrigger }) => {
  const state = usePublicActivity(groupId ? `/v0/market-groups/${groupId}/bets` : '', refreshTrigger);
  return <PublicTradingSummary {...state} />;
};
