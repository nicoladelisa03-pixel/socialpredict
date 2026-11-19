import React from 'react';
import { Link } from 'react-router-dom';
import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';
import { useMarketDetails } from '../../hooks/useMarketDetails';
import { useAuth } from '../../helpers/AuthContent';
import LoadingSpinner from '../../components/loaders/LoadingSpinner';
import { useDocumentMeta } from '../../hooks/useDocumentMeta';
import { DOMAIN_URL } from '../../config';

const MarketDetails = () => {
  const { username, usertype, moderatorStatus } = useAuth();
  const { details, availability, isLoggedIn, token, refetchData, currentProbability } =
    useMarketDetails();

  useDocumentMeta({
    title: details?.market?.questionTitle
      ? `${details.market.questionTitle} | SocialPredict`
      : 'SocialPredict',
    description: details
      ? `${Math.round(currentProbability * 100)}% probability · created by ${details.creator?.username || details.market.creatorUsername}`
      : undefined,
    url: details?.market?.id ? `${DOMAIN_URL}/markets/${details.market.id}` : undefined,
  });

  if (availability === 'not-found') {
    return (
      <div className='flex flex-col text-white px-4 py-8'>
        <div className='rounded-lg border border-gray-600 bg-gray-800 p-4 md:p-6' role='status'>
          <h1 className='text-2xl font-bold mb-2'>Market not found</h1>
          <p className='text-gray-300 mb-4'>This market is not available.</p>
          <Link to='/markets' className='text-blue-400 hover:text-blue-300'>
            All markets
          </Link>
        </div>
      </div>
    );
  }

  if (availability === 'error') {
    return (
      <div className='flex flex-col text-white px-4 py-8'>
        <div className='rounded-lg border border-red-600 bg-red-900/50 p-4' role='alert'>
          <p className='text-red-300 mb-3'>This market could not be loaded.</p>
          <button
            type='button'
            onClick={refetchData}
            className='px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors'
          >
            Try again
          </button>
        </div>
      </div>
    );
  }

  if (!details) {
    return <LoadingSpinner />;
  }

  return (
    <div className='flex flex-col min-h-screen'>
      <div className='flex-grow overflow-y-auto'>
        <MarketDetailsTable
          market={details.market}
          creator={details.creator}
          numUsers={details.numUsers}
          totalVolume={details.totalVolume}
          marketDust={details.marketDust || 0}
          currentProbability={currentProbability}
          probabilityChanges={details.probabilityChanges}
          freshness={details.freshness}
          descriptionAmendments={details.descriptionAmendments}
          marketId={details.market.id}
          username={username}
          usertype={usertype}
          moderatorStatus={moderatorStatus}
          isLoggedIn={isLoggedIn}
          token={token}
          refetchData={refetchData}
        />
      </div>
    </div>
  );
};

export default MarketDetails;
