/** @vitest-environment jsdom */
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { renderHook, cleanup } from '@testing-library/react';
import { SITE_DESCRIPTION, SITE_NAME } from '../config';
import { useDocumentMeta } from '../hooks/useDocumentMeta';

function getMeta(property) {
  const el = document.querySelector(`meta[property="${property}"]`);
  return el ? el.getAttribute('content') : null;
}

function getNameMeta(name) {
  const el = document.querySelector(`meta[name="${name}"]`);
  return el ? el.getAttribute('content') : null;
}

beforeEach(() => {
  // Remove any og: tags left from prior tests
  document.querySelectorAll('meta[property^="og:"]').forEach((el) => el.remove());
  document.querySelectorAll('meta[name^="twitter:"]').forEach((el) => el.remove());
  document.title = SITE_NAME;
});

afterEach(cleanup);

describe('useDocumentMeta', () => {
  it('sets document.title to the provided title', () => {
    renderHook(() => useDocumentMeta({ title: `Will it rain? — ${SITE_NAME}` }));
    expect(document.title).toBe(`Will it rain? — ${SITE_NAME}`);
  });

  it('sets og:title meta tag', () => {
    renderHook(() => useDocumentMeta({ title: 'My market' }));
    expect(getMeta('og:title')).toBe('My market');
  });

  it('sets og:description meta tag', () => {
    renderHook(() =>
      useDocumentMeta({ title: 'Test', description: '75% probability · created by alice' })
    );
    expect(getMeta('og:description')).toBe('75% probability · created by alice');
  });

  it('sets og:type to website', () => {
    renderHook(() => useDocumentMeta({ title: 'Test' }));
    expect(getMeta('og:type')).toBe('website');
  });

  it('resets document.title on unmount', () => {
    document.title = 'Original Title';
    const { unmount } = renderHook(() =>
      useDocumentMeta({ title: 'Temporary Title' })
    );
    expect(document.title).toBe('Temporary Title');
    unmount();
    expect(document.title).toBe('Original Title');
  });

  it('resets og:title to site default on unmount', () => {
    const { unmount } = renderHook(() =>
      useDocumentMeta({ title: 'Market question?' })
    );
    unmount();
    expect(getMeta('og:title')).toBe(SITE_NAME);
  });

  it('uses site default description when none provided', () => {
    renderHook(() => useDocumentMeta({ title: 'Test' }));
    expect(getMeta('og:description')).toBe(SITE_DESCRIPTION);
  });

  it('sets image and twitter card metadata', () => {
    renderHook(() =>
      useDocumentMeta({
        title: 'Market title',
        description: 'Market description',
        image: 'https://example.test/share.png',
      })
    );

    expect(getMeta('og:image')).toBe('https://example.test/share.png');
    expect(getMeta('og:site_name')).toBe(SITE_NAME);
    expect(getNameMeta('twitter:card')).toBe('summary_large_image');
    expect(getNameMeta('twitter:title')).toBe('Market title');
    expect(getNameMeta('twitter:description')).toBe('Market description');
    expect(getNameMeta('twitter:image')).toBe('https://example.test/share.png');
  });
});
