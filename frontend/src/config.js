export const DOMAIN_URL = (typeof window !== 'undefined' && window.__ENV__ && window.__ENV__.DOMAIN_URL) ? window.__ENV__.DOMAIN_URL : 'http://localhost';
export const API_URL = (typeof window !== 'undefined' && window.__ENV__ && window.__ENV__.API_URL) ? window.__ENV__.API_URL : 'http://localhost';

// Temporary product name until a founder chooses a permanent brand.
export const SITE_NAME = 'Startup Predictions';
export const SITE_DESCRIPTION = 'Prediction markets about startups, traded with virtual credits.';
