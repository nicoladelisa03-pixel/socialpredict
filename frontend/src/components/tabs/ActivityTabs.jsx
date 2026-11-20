import React from 'react';
import SiteTabs from './SiteTabs';
import { PublicMarketActivity } from '../layouts/activity/PublicTradingActivity';

const ActivityTabs = ({ marketId, refreshTrigger }) => {
    const tabsData = [
        { label: 'Activity', content: <PublicMarketActivity marketId={marketId} refreshTrigger={refreshTrigger} /> },
        { label: 'Comments', content: <div>Comments Go here...</div> },
    ];

    return <SiteTabs tabs={tabsData} />;
};

export default ActivityTabs;
