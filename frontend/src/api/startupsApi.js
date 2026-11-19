import { apiRequest, authenticatedApiRequest } from './httpClient';

const startupReasonMessages = {
  AUTHORIZATION_DENIED: 'You need an admin account to manage startups.',
  INVALID_TOKEN: 'Sign in again to manage startups.',
  VALIDATION_FAILED: 'Check the startup fields and try again.',
  NOT_FOUND: 'That startup or market could not be found.',
  INVALID_REQUEST: 'Check the form and try again.',
};

export const listStartups = () => apiRequest('/v0/startups', {
  fallbackMessage: 'Unable to load startups.',
});

export const getStartup = (slug) => apiRequest(`/v0/startups/${encodeURIComponent(slug)}`, {
  fallbackMessage: 'Unable to load this startup.',
});

export const getStartupMarkets = (slug) => apiRequest(`/v0/startups/${encodeURIComponent(slug)}/markets`, {
  fallbackMessage: 'Unable to load this startup\'s markets.',
});

export const getStartupFunding = (slug) => apiRequest(`/v0/startups/${encodeURIComponent(slug)}/funding`, {
  fallbackMessage: 'Unable to load funding history.',
});

export const createAdminStartup = (startup) => authenticatedApiRequest('/v0/admin/startups', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(startup),
  reasonMessages: startupReasonMessages,
  fallbackMessage: 'Unable to create this startup.',
});

export const updateAdminStartup = (slug, startup) => authenticatedApiRequest(`/v0/admin/startups/${encodeURIComponent(slug)}`, {
  method: 'PATCH',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(startup),
  reasonMessages: startupReasonMessages,
  fallbackMessage: 'Unable to update this startup.',
});

export const listAdminStartupMarkets = () => authenticatedApiRequest('/v0/admin/startup-markets', {
  reasonMessages: startupReasonMessages,
  fallbackMessage: 'Unable to load markets for startup linking.',
});

export const createAdminStartupFunding = (slug, round) => authenticatedApiRequest(`/v0/admin/startups/${encodeURIComponent(slug)}/funding`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(round),
  reasonMessages: startupReasonMessages,
  fallbackMessage: 'Unable to add this funding round.',
});

export const updateAdminStartupFunding = (slug, entryId, round) => authenticatedApiRequest(`/v0/admin/startups/${encodeURIComponent(slug)}/funding/${encodeURIComponent(entryId)}`, {
  method: 'PATCH',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(round),
  reasonMessages: startupReasonMessages,
  fallbackMessage: 'Unable to update this funding round.',
});

export const setAdminMarketStartup = (marketId, startupSlug) => authenticatedApiRequest(`/v0/admin/markets/${encodeURIComponent(marketId)}/startup`, {
  method: 'PATCH',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ startupSlug }),
  reasonMessages: startupReasonMessages,
  fallbackMessage: 'Unable to update this market\'s startup.',
});
