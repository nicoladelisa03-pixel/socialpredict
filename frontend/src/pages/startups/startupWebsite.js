export const safeStartupWebsite = (value) => {
  if (typeof value !== 'string' || value.trim() === '') {
    return '';
  }
  try {
    const url = new URL(value.trim());
    if (url.protocol !== 'http:' && url.protocol !== 'https:') {
      return '';
    }
    if (!url.hostname) {
      return '';
    }
    return url.toString();
  } catch {
    return '';
  }
};

const testStartupSlugs = new Set(['step3-verify', 'step8-admin']);

export const isTestStartup = (slug) => testStartupSlugs.has(slug);

export const startupWebsiteLabel = (url) => {
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return 'Website';
  }
};
